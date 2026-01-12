// single level routes
const Home = () => "/"
const Features = () => "/features"
const UseCases = () => "/use-cases"
const TermsAndConditions = () => "/terms-and-conditions"
const PrivacyPolicy = () => "/privacy-policy"
const Profile = () => "/profile"
const SignIn = () => "/sign-in"
const SignUp = () => "/sign-up"
const ForgotPassword = () => "/forgot-password"

// Docs
interface DocsRoutes {
  Tutorials: () => string
}
const Docs: (() => string) & DocsRoutes = () => "/docs"
Docs.Tutorials = () => `${Docs()}/tutorials`

// admin routes
interface AdminRoutes {
  Users: () => string
  Vouchers: () => string
  System: () => string
  Invoices: () => string
  Payments: () => string
  Emails: () => string
  Workflows: () => string
}

const Admin: (() => string) & AdminRoutes = () => "/admin"
Admin.Users = () => `${Admin()}/users`
Admin.Vouchers = () => `${Admin()}/vouchers`
Admin.System = () => `${Admin()}/system`
Admin.Invoices = () => `${Admin()}/invoices`
Admin.Payments = () => `${Admin()}/payments`
Admin.Emails = () => `${Admin()}/emails`
Admin.Workflows = () => `${Admin()}/workflows`

// dashboard routes
interface DashboardRoutes {
  Clusters: (() => string) & { Deploy: () => string }
  MyNodes: (() => string) & { Explorer: () => string }
  SshKeys: () => string
  Funds: () => string
  BillingHistory: () => string
  Payments: () => string
  Vouchers: () => string
}

const Dashboard: (() => string) & DashboardRoutes = () => "/dashboard"
Dashboard.Clusters = (() => `${Dashboard()}/clusters`) as DashboardRoutes["Clusters"]
Dashboard.Clusters.Deploy = () => `${Dashboard.Clusters()}/deploy`
Dashboard.MyNodes = (() => `${Dashboard()}/my-nodes`) as DashboardRoutes["MyNodes"]
Dashboard.MyNodes.Explorer = () => `${Dashboard.MyNodes()}/explorer`
Dashboard.SshKeys = () => `${Dashboard()}/ssh-keys`
Dashboard.Funds = () => `${Dashboard()}/funds`
Dashboard.BillingHistory = () => `${Dashboard()}/billing-history`
Dashboard.Payments = () => `${Dashboard()}/payments`
Dashboard.Vouchers = () => `${Dashboard()}/vouchers`

export const ROUTES = {
  Home,
  Features,
  Docs,
  UseCases,
  TermsAndConditions,
  PrivacyPolicy,
  Profile,
  SignIn,
  SignUp,
  ForgotPassword,
  Admin,
  Dashboard,
}
