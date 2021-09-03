/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import Dashboard from "../pages/dashboard/Dashboard";
import Flavors from "../pages/flavors/Flavors";
import CreateFlavors from "../pages/flavors/CreateFlavors";
import Floatingips from "../pages/floatingips/Floatingips";
import CreateFloatingips from "../pages/floatingips/CreateFloatingips";
import Gateways from "../pages/gateways/Gateways";
import Images from "../pages/images/Images";
import CreateImages from "../pages/images/CreateImages";
import Instances from "../pages/instances/Instances";
import ModifyInstances from "../pages/instances/ModifyInstances";
import Keys from "../pages/keys/Keys";
import Login from "../pages/login/Login";
import Openshifts from "../pages/openshifts/Openshifts";
import ModifyOpenshifts from "../pages/openshifts/ModifyOpenshifts";

import Orgs from "../pages/orgs/Orgs";
import CreateOrg from "../pages/orgs/CreateOrg";
import ModifyOrg from "../pages/orgs/ModifyOrg";

import PageNotFound from "../pages/PageNotFound";
import Registrys from "../pages/registrys/Registrys";
import ModifyRegistrys from "../pages/registrys/ModifyRegistrys";
import Secgroups from "../pages/secgroups/Secgroups";
import Secrules from "../pages/secrules/Secrules";
import ModifySecrules from "../pages/secrules/ModifySecrules";
import Subnets from "../pages/subnets/Subnets";
import Users from "../pages/users/Users";
import ModifyUser from "../pages/users/ModifyUser";
import Hypers from "../pages/hypers/Hypers";
import ModifySubnets from "../pages/subnets/ModifySubnets";
import ModifyGateways from "../pages/gateways/ModifyGateways";
import ModifySecgroups from "../pages/secgroups/ModifySecgroups";
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
    title: "User",
    icon: "user",
    isShow: true,
    item: "auth",
  },
  // {
  //   path: "/users/new",
  //   component: CreateUser,
  //   exact: true,
  //   isShow: false,
  //   item: "auth",
  // },
  {
    path: "/users/new/:id?",
    component: ModifyUser,
    exact: true,
    isShow: false,
    item: "auth",
  },
  {
    path: "/orgs",
    component: Orgs,
    title: "Organization",
    icon: "team",
    exact: true,
    isShow: true,
    item: "auth",
  },
  {
    path: "/orgs/new",
    component: CreateOrg,
    exact: true,
    isShow: false,
    item: "auth",
  },
  {
    path: "/orgs/new/:id?",
    component: ModifyOrg,
    exact: true,
    isShow: false,
    item: "auth",
  },
  {
    path: "/keys",
    component: Keys,
    title: "Key",
    exact: true,
    isShow: true,
    item: "auth",
  },
  {
    path: "/instances",
    component: Instances,
    title: "Instance",
    exact: true,
    isShow: true,
    item: "compute",
  },
  {
    path: "/instances/new/:id?",
    component: ModifyInstances,
    isShow: false,
    item: "compute",
  },
  {
    path: "/flavors",
    component: Flavors,
    title: "Flavor",
    exact: true,
    isShow: true,
    item: "compute",
  },
  {
    path: "/flavors/new",
    component: CreateFlavors,
    isShow: false,
    item: "compute",
  },
  {
    path: "/images",
    component: Images,
    title: "Image",
    exact: true,
    isShow: true,
    item: "compute",
  },
  {
    path: "/images/new",
    component: CreateImages,
    isShow: false,
    item: "compute",
  },
  {
    path: "/openshifts",
    component: Openshifts,
    title: "Openshift",
    exact: true,
    isShow: true,
    item: "platform",
  },
  {
    path: "/openshifts/new/:id?",
    component: ModifyOpenshifts,
    isShow: false,
    item: "platform",
  },
  {
    path: "/registrys",
    component: Registrys,
    title: "Registry",
    exact: true,
    isShow: true,
    item: "platform",
  },
  {
    path: "/registrys/new/:id?",
    component: ModifyRegistrys,
    isShow: false,
    item: "platform",
  },

  {
    path: "/subnets",
    component: Subnets,
    title: "Subnet",
    exact: true,
    isShow: true,
    item: "network",
  },
  {
    path: "/subnets/new/:id?",
    component: ModifySubnets,
    isShow: false,
    item: "network",
  },
  {
    path: "/floatingips",
    component: Floatingips,
    title: "FloatingIp",
    exact: true,
    isShow: true,
    item: "network",
  },
  {
    path: "/floatingips/new/:id?",
    component: CreateFloatingips,
    isShow: false,
    item: "network",
  },
  {
    path: "/gateways",
    component: Gateways,
    title: "Gateway",
    exact: true,
    isShow: true,
    item: "network",
  },
  {
    path: "/gateways/new/:id?",
    component: ModifyGateways,
    isShow: false,
    item: "network",
  },
  {
    path: "/secgroups",
    title: "SecurityGroup",
    component: Secgroups,
    exact: true,
    isShow: true,
    item: "network",
  },
  {
    path: "/secgroups/new/:id?",
    component: ModifySecgroups,
    isShow: false,
    item: "network",
  },

  {
    path: "/secgroups/:id/secrules",
    component: Secrules,
    exact: true,
    isShow: false,
    item: "network",
  },
  {
    path: "/secgroups/:id/secrules/new/:id?",
    component: ModifySecrules,
    isShow: false,
    item: "network",
  },
  {
    path: "/hypers",
    component: Hypers,
    title: "Hypervisors",
    exact: true,
    isShow: true,
    item: "admin",
  },
];
// export default (InitRoutes, mainRoutes);
