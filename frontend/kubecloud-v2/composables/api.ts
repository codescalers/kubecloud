import { Lock as MutexLock } from "async-await-mutex-lock"
import axios, { AxiosError } from "axios"
import {
  AdminApi,
  DeploymentsApi,
  InvoicesApi,
  NodesApi,
  NotificationsApi,
  TwinsApi,
  UsersApi,
  WorkflowApi,
} from "~/generated/api"

export const useApi = createGlobalState(() => {
  const mu = new MutexLock()
  async function awaitLock() {
    if (mu.isAcquired()) {
      await mu.acquire()
      mu.release()
    }
  }

  const instance = axios.create({
    baseURL: useRuntimeConfig().public.apiBasePath,
  })

  const admin = new AdminApi(undefined, undefined, instance)
  const deployments = new DeploymentsApi(undefined, undefined, instance)
  const invoices = new InvoicesApi(undefined, undefined, instance)
  const nodes = new NodesApi(undefined, undefined, instance)
  const notifications = new NotificationsApi(undefined, undefined, instance)
  const twins = new TwinsApi(undefined, undefined, instance)
  const users = new UsersApi(undefined, undefined, instance)
  const workflow = new WorkflowApi(undefined, undefined, instance)

  const { accessToken, refreshToken } = useTokens()

  instance.interceptors.request.use(async (config) => {
    if (config.unauthenticated) {
      return config
    }

    // await lock to get access to accessToken
    await awaitLock()
    config.headers.Authorization = `Bearer ${accessToken.value}`

    return config
  })

  instance.interceptors.response.use(
    response => response,
    async (error) => {
      if (
        error instanceof AxiosError
        && error.response
        && [401, 403].includes(error.response.status)
        && error.config
        && "Authorization" in error.config.headers
      ) {
        if (!mu.isAcquired()) {
          await mu.acquire()
          const { data } = await users.refreshToken(
            { refresh_token: refreshToken.value },
            { unauthenticated: true },
          )

          accessToken.value = data.data?.access_token ?? ""
          mu.release()
        }
        else {
          await awaitLock()
        }

        return instance(error.config)
      }

      return Promise.reject(error)
    },
  )

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
      workflow,
    ),
  }
})

class ApiHelpers {
  private readonly twinToAccountId: { [twinId: number]: string }
  private readonly clusterToKubeconfig: { [clusterId: string]: string }

  constructor(
    private readonly admin: AdminApi,
    private readonly deployments: DeploymentsApi,
    private readonly invoices: InvoicesApi,
    private readonly nodes: NodesApi,
    private readonly notifications: NotificationsApi,
    private readonly twins: TwinsApi,
    private readonly users: UsersApi,
    private readonly workflow: WorkflowApi,
  ) {
    this.twinToAccountId = {}
    this.clusterToKubeconfig = {}
  }

  public async awaitWorkflowCompletion(workflowId: string): Promise<boolean> {
    const { data } = await this.workflow.getWorkflowStatus(workflowId, {
      unauthenticated: true,
    })

    if (data.data === "failed") {
      return false
    }

    if (data.data === "completed") {
      return true
    }

    await new Promise(res => setTimeout(res, 2_000))
    return this.awaitWorkflowCompletion(workflowId)
  }

  public async getAccountId(twinId: number): Promise<string> {
    if (twinId in this.twinToAccountId) {
      return this.twinToAccountId[twinId]!
    }

    // TODO: handle error
    const { data } = await this.twins.twinsTwinIdAccountGet(twinId)
    const accountId = data.data?.account_id ?? ""
    this.twinToAccountId[twinId] = accountId
    return accountId
  }

  public async getKubeconfig(clusterName: string): Promise<string> {
    if (clusterName in this.clusterToKubeconfig) {
      return this.clusterToKubeconfig[clusterName]!
    }

    const { data } = await this.deployments.deploymentsNameKubeconfigGet(clusterName)
    this.clusterToKubeconfig[clusterName] = data.data?.kubeconfig ?? ""
    return data.data?.kubeconfig ?? ""
  }
}
