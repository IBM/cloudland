import Dashboard from "../pages/dashboard/Dashboard";
import Flavors from "../pages/flavors/Flavors";
import Floatingips from "../pages/floatingips/Floatingips";
import Gateways from "../pages/gateways/Gateways";
import Images from "../pages/images/Images";
import Instances from "../pages/instances/Instances";
import ModifyInstances from "../pages/instances/ModifyInstances";
import Keys from "../pages/keys/Keys";
import Login from "../pages/login/Login";
import Openshifts from "../pages/openshifts/Openshifts";
import Orgs from "../pages/orgs/Orgs";

import PageNotFound from "../pages/PageNotFound";
import Registrys from "../pages/registrys/Registrys";
import ModifyRegistrys from "../pages/registrys/ModifyRegistrys";
import Secgroups from "../pages/secgroups/Secgroups";
import Secrules from "../pages/secgroups/secrules/Secrules";
import Subnets from "../pages/subnets/Subnets";
import Users from "../pages/users/Users";
import Hypers from "../pages/hypers/Hypers";
export const InitRoutes = [
  {
    path: "/login",
    component: Login,
    //exact: true,
  },
  {
    path: "/404",
    component: PageNotFound,
  },
];
export const mainRoutes = [
  {
    path: "/dashboard",
    component: Dashboard,
    title: "Dashboard",
    //icon:''
  },
  {
    path: "/users",
    component: Users,
    exact: true,
    title: "Users",
    icon: "user",
    isShow: true,
    item: "auth",
  },
  {
    path: "/orgs",
    component: Orgs,
    title: "Organizations",
    icon: "team",
    exact: true,
    isShow: true,
    item: "auth",
  },
  {
    path: "/keys",
    component: Keys,
    exact: true,
    isShow: true,
    item: "auth",
    title: "Keys",
  },
  {
    path: "/instances",
    component: Instances,
    title: "Instances",
    exact: true,
    isShow: true,
    item: "compute",
  },
  {
    path: "/instances/new/:id?",
    component: ModifyInstances,
    exact: true,
    isShow: false,
    item: "compute",
  },
  {
    path: "/flavors",
    component: Flavors,
    title: "Flavors",
    exact: true,
    isShow: true,
    item: "compute",
  },
  {
    path: "/images",
    component: Images,
    title: "Images",
    exact: true,
    isShow: true,
    item: "compute",
  },
  {
    path: "/registrys",
    component: Registrys,
    title: "Registry",
    exact: true,
    isShow: true,
    item: "compute",
  },
  {
    path: "/registrys/new/:id?",
    component: ModifyRegistrys,
    exact: true,
    isShow: false,
    item: "compute",
  },
  {
    path: "/openshifts",
    component: Openshifts,
    title: "Openshifts",
    exact: true,
    isShow: true,
    item: "platform",
  },
  {
    path: "/subnets",
    component: Subnets,
    title: "Subnets",
    exact: true,
    isShow: true,
    item: "network",
  },
  {
    path: "/floatingips",
    component: Floatingips,
    title: "FloatingIps",
    exact: true,
    isShow: true,
    item: "network",
  },
  {
    path: "/gateways",
    component: Gateways,
    title: "Gateways",
    exact: true,
    isShow: true,
    item: "network",
  },
  {
    path: "/secgroups",
    title: "SecurityGroups",
    component: Secgroups,
    exact: true,
    isShow: true,
    item: "network",
  },
  {
    path: "/secgroups/:id/secrules",
    component: Secrules,
    //exact: true,
    isShow: false,
    item: "network",
  },
  {
    path: "/hypers",
    component: Hypers,
    title: "Hypers",
    exact: true,
    isShow: true,
    item: "admin",
  },
];
