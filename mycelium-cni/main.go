// Copyright 2015 CNI authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"crypto/rand"
	"fmt"
	"net"
	"os"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ns"
	bv "github.com/containernetworking/plugins/pkg/utils/buildversion"
	"github.com/containernetworking/plugins/plugins/ipam/host-local/backend/allocator"
	"github.com/containernetworking/plugins/plugins/ipam/host-local/backend/disk"
	"github.com/threefoldtech/zosbase/pkg/netlight/ifaceutil"
	"github.com/threefoldtech/zosbase/pkg/netlight/resource"
	"github.com/vishvananda/netlink"
)

const myceliumBr = "mycelium-br"

func main() {
	skel.PluginMainFuncs(skel.CNIFuncs{
		Add:   cmdAdd,
		Check: cmdCheck,
		Del:   cmdDel,
		/* FIXME GC */
		/* FIXME Status */
	}, version.All, bv.BuildString("host-local"))
}

func cmdCheck(args *skel.CmdArgs) error {
	ipamConf, _, err := allocator.LoadIPAMConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}

	// Look to see if there is at least one IP address allocated to the container
	// in the data dir, irrespective of what that address actually is
	store, err := disk.New(ipamConf.Name, ipamConf.DataDir)
	if err != nil {
		return err
	}
	defer store.Close()

	containerIPFound := store.FindByID(args.ContainerID, args.IfName)
	if !containerIPFound {
		return fmt.Errorf("host-local: Failed to find address added by container %v", args.ContainerID)
	}

	return nil
}
func RandomMyceliumIPSeed() ([]byte, error) {
	key := make([]byte, 6)
	_, err := rand.Read(key)
	return key, err
}

func cmdAdd(args *skel.CmdArgs) error {
	netSeed, err := os.ReadFile("/etc/netseed")
	if err != nil {
		return err
	}

	inspect, err := resource.InspectMycelium(netSeed)
	if err != nil {
		return err
	}

	vmSeed, err := RandomMyceliumIPSeed()
	if err != nil {
		return fmt.Errorf("failed to generate random seed for VM: %v", err)
	}

	ip, gw, err := inspect.IPFor(vmSeed)
	if err != nil {
		return err
	}

	containerNetns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer containerNetns.Close()

	vethName := fmt.Sprintf("veth%s", args.ContainerID)
	vethName = vethName[0:12]
	vethLink, err := ifaceutil.MakeVethPair(vethName, myceliumBr, 1500, "")
	if err != nil {
		return fmt.Errorf("failed to create veth pair: %v", err)
	}

	// Move container veth to container netns
	if err := netlink.LinkSetNsFd(vethLink, int(containerNetns.Fd())); err != nil {
		return fmt.Errorf("failed to move veth to container: %v", err)
	}

	// Setup in container netns
	err = containerNetns.Do(func(_ ns.NetNS) error {
		contVeth, err := netlink.LinkByName(vethName)
		if err != nil {
			return fmt.Errorf("failed to get link by name: %s, %v ", vethName, err)
		}

		addr, _ := netlink.ParseAddr(ip.String())
		if err := netlink.AddrAdd(contVeth, addr); err != nil {
			return fmt.Errorf("failed to add address to: %s, %v ", vethName, err)
		}

		if err := netlink.LinkSetUp(contVeth); err != nil {
			return fmt.Errorf("failed to set link up : %s, %v ", vethName, err)
		}

		dst, err := netlink.ParseIPNet("400::/7")
		if err != nil {
			return fmt.Errorf("error parsing route destination: %v\n", err)
		}

		// Create the route
		route := &netlink.Route{
			Dst: dst,
			Gw:  gw.IP,
		}
		return netlink.RouteAdd(route)
	})
	if err != nil {
		return fmt.Errorf("failed to configure veth in container: %v", err)
	}

	// Return result
	result := &current.Result{
		CNIVersion: "0.4.0",
		Interfaces: []*current.Interface{
			{
				Name:    vethName,
				Sandbox: args.Netns,
			},
		},
		IPs: []*current.IPConfig{
			{
				Address: net.IPNet{
					IP:   ip.IP,
					Mask: net.CIDRMask(64, 128),
				},
				Gateway: gw.IP,
			},
		},
	}

	return types.PrintResult(result, "0.4.0")
}

func cmdDel(args *skel.CmdArgs) error {
	return nil
}
