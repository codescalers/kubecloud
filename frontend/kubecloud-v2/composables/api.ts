import axios from "axios"
import {
  AdminApi,
  DeploymentsApi,
  InvoicesApi,
  NodesApi,
  NotificationsApi,
  TwinsApi,
  UsersApi,
  WorkflowApi,
} from "../generated/api"

class ApiHelpers {
  constructor(
    private readonly admin: AdminApi,
    private readonly deployments: DeploymentsApi,
    private readonly invoices: InvoicesApi,
    private readonly nodes: NodesApi,
    private readonly notifications: NotificationsApi,
    private readonly twins: TwinsApi,
    private readonly users: UsersApi,
    private readonly workflow: WorkflowApi
  ) {}

  public async awaitWorkflowCompletion(
    workflowId: string,
    deplay: number = 5_000
  ): Promise<boolean> {
    const { data } = await this.workflow.getWorkflowStatus(workflowId, {
      _flags: { unauthenticated: true },
    })

    console.log("state", data.data)

    if (data.data === "failed") {
      return false
    }

    if (data.data === "completed") {
      return true
    }

    await new Promise((res) => setTimeout(res, deplay))
    return this.awaitWorkflowCompletion(workflowId, deplay)
  }
}

export const useApi = createGlobalState(() => {
  const config = useRuntimeConfig()
  const { apiBasePath } = config.public

  const accessToken = useLocalStorage<string>("accessToken", "")
  // const refreshToken = useLocalStorage<string>("refreshToken", "")

  const instance = axios.create({
    baseURL: apiBasePath,
  })

  instance.interceptors.request.use((config) => {
    if (config._flags?.unauthenticated) {
      return config
    }

    // await lock to get access to accessToken
    config.headers.Authorization = `Bearer ${accessToken.value}`

    return config
  })

  /*
  instance.interceptors.response.use(
    (response) => response,
    (error: AxiosError) => {
      if (error.response?.status === 401) {
        // local access token expired, clear it
        // request new access token
        // update tokens
        // unlock access to accessToken
        // retry request
      }

      // if (error.response?.status === 401) {
      //   // await lock to get access to refreshToken
      //   const newAccessToken = await refreshToken.value
      //   if (newAccessToken) {
      //     accessToken.value = newAccessToken
      //   }
      // }

      return Promise.reject(error)
    }
  ) */

  // instance.interceptors.request.use((config) => {
  //   console.log("[request] _internalFlags", config._internalFlags)
  //   return config
  // })

  // instance.interceptors.response.use(
  //   (response) => {
  //     console.log("[response] _internalFlags", response.config._internalFlags)
  //     return response
  //   },
  //   (error: AxiosError) => {
  //     console.log(error.config?._internalFlags)
  //     console.log("[response] error", error)
  //     return Promise.reject(error)
  //   }
  // )

  // const config = new Configuration({
  //   basePath: "https://staging.myceliumcloud.tf/api/v1",
  // })

  const admin = new AdminApi(undefined, undefined, instance)
  const deployments = new DeploymentsApi(undefined, undefined, instance)
  const invoices = new InvoicesApi(undefined, undefined, instance)
  const nodes = new NodesApi(undefined, undefined, instance)
  const notifications = new NotificationsApi(undefined, undefined, instance)
  const twins = new TwinsApi(undefined, undefined, instance)
  const users = new UsersApi(undefined, undefined, instance)
  const workflow = new WorkflowApi(undefined, undefined, instance)

  return {
    admin,
    deployments,
    invoices,
    nodes,
    notifications,
    twins,
    users,
    workflow,
    helpers: new ApiHelpers(
      admin,
      deployments,
      invoices,
      nodes,
      notifications,
      twins,
      users,
      workflow
    ),
  }
})
