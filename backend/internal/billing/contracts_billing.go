package billing

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/graphql"
)

var ErrorEventsNotFound = fmt.Errorf("could not find any events")

type ContractBillReports struct {
	Reports []Report `json:"contractBillReports"`
}

type Report struct {
	ContractID   string `json:"contractID"`
	Timestamp    string `json:"timestamp"`
	AmountBilled string `json:"amountBilled"`
}

// ListContractBillReportsPerMonth returns bill reports for contract ID month ago
func ListContractBillReportsPerMonth(graphqlClient graphql.GraphQl, contractID uint64, currentTime time.Time) (ContractBillReports, error) {
	monthAgo := currentTime.AddDate(0, -1, 0)

	options := fmt.Sprintf(`(where: {contractID_eq: %v, timestamp_lte: %v, timestamp_gte: %v}, orderBy: id_ASC)`, contractID, currentTime.Unix(), monthAgo.Unix())
	billingReportsCount, err := graphqlClient.GetItemTotalCount("contractBillReports", options)
	if err != nil {
		return ContractBillReports{}, err
	}
	billingReportsData, err := graphqlClient.Query(fmt.Sprintf(`query MyQuery($billingReportsCount: Int!){
            contractBillReports(where: {contractID_eq: %v, timestamp_lte: %v, timestamp_gte: %v}, limit: $billingReportsCount) {
              contractID
              timestamp
              amountBilled
            }
          }`, contractID, currentTime.Unix(), monthAgo.Unix()),
		map[string]interface{}{
			"billingReportsCount": billingReportsCount,
		})
	if err != nil {
		return ContractBillReports{}, err
	}

	billingReports, err := json.Marshal(billingReportsData)
	if err != nil {
		return ContractBillReports{}, err
	}

	var reports ContractBillReports
	err = json.Unmarshal(billingReports, &reports)
	if err != nil {
		return ContractBillReports{}, err
	}

	return reports, nil
}

// TODO: check returned float or int
func AmountBilledPerMonth(reports ContractBillReports) (uint64, error) {
	var totalAmount uint64
	for _, report := range reports.Reports {
		amount, err := strconv.ParseInt(report.AmountBilled, 10, 64)
		if err != nil {
			return 0, err
		}

		totalAmount += uint64(amount)
	}

	return totalAmount, nil
}

type Events struct {
	Events []Event `json:"events"`
}

type Event struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Block struct {
		Height    uint64 `json:"height"`
		Timestamp string `json:"timestamp"`
	} `json:"block"`
}

func GetRentContractCancellationDate(graphql graphql.GraphQl, contractID uint64) (time.Time, error) {
	options := fmt.Sprintf(`(where: {args_jsonContains: "{\"contractId\": \"%v\"}", name_contains: "SmartContractModule.RentContractCanceled"}, orderBy: block_timestamp_DESC, limit: 1,)`, contractID)

	eventsData, err := graphql.Query(fmt.Sprintf(`query getEvents{
            events%v {
              name
              id
              block {
                height
                timestamp
              }
            }
          }`, options),
		map[string]interface{}{})
	if err != nil {
		return time.Time{}, err
	}

	eventsJSONData, err := json.Marshal(eventsData)
	if err != nil {
		return time.Time{}, err
	}

	var listEvents Events
	err = json.Unmarshal(eventsJSONData, &listEvents)
	if err != nil {
		return time.Time{}, err
	}

	if len(listEvents.Events) == 0 {
		return time.Time{}, errors.Wrapf(ErrorEventsNotFound, "no events found for contract %d", contractID)
	}

	parsedTime, err := time.Parse(time.RFC3339, listEvents.Events[0].Block.Timestamp)
	if err != nil {
		return time.Time{}, errors.Wrapf(err, "failed to parse timestamp for contract %d", contractID)
	}

	return parsedTime, nil
}
