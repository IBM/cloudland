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
import CreateKey from "../pages/keys/CreateKey";
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
import CreateUser from "../pages/users/CreateUser";
import Hypers from "../pages/hypers/Hypers";
import ModifySubnets from "../pages/subnets/ModifySubnets";
import ModifyGateways from "../pages/gateways/ModifyGateways";
import ModifySecgroups from "../pages/secgroups/ModifySecgroups";
import Registers from "../pages/registers/Registers";
export const InitRoutes = [
  {
    path: "/login",
    component: Login,
    //exact: true,
  },
  {
    path: "/registers",
    component: Registers,
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
    path: "/users/new",
    component: CreateUser,
    exact: true,
    isShow: false,
    item: "auth",
  },
  {
    path: "/users/:id?",
    component: ModifyUser,
    exact: true,
    isShow: false,
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
    path: "/orgs/new",
    component: CreateOrg,
    exact: true,
    isShow: false,
    item: "auth",
  },
  {
    path: "/orgs/:id?",
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
    path: "/keys/new",
    component: CreateKey,
    exact: true,
    isShow: false,
    item: "auth",
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
    path: "/instances/new",
    component: ModifyInstances,
    isShow: false,
    item: "compute",
  },
  {
    path: "/instances/:id?",
    component: ModifyInstances,
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
    path: "/flavors/new",
    component: CreateFlavors,
    isShow: false,
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
    path: "/openshifts/new",
    component: ModifyOpenshifts,
    isShow: false,
    item: "platform",
  },
  {
    path: "/openshifts/:id?",
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
    path: "/registrys/new",
    component: ModifyRegistrys,
    isShow: false,
    item: "platform",
  },
  {
    path: "/registrys/:id?",
    component: ModifyRegistrys,
    isShow: false,
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
    path: "/subnets/new",
    component: ModifySubnets,
    isShow: false,
    item: "network",
  },
  {
    path: "/subnets/:id?",
    component: ModifySubnets,
    isShow: false,
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
    path: "/floatingips/new",
    component: CreateFloatingips,
    isShow: false,
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
    path: "/gateways/new",
    component: ModifyGateways,
    isShow: false,
    item: "network",
  },
  {
    path: "/gateways/:id?",
    component: ModifyGateways,
    isShow: false,
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
    path: "/secgroups/new",
    component: ModifySecgroups,
    isShow: false,
    item: "network",
  },
  {
    path: "/secgroups/:id?",
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
    path: "/secgroups/:id/secrules/new",
    component: ModifySecrules,
    isShow: false,
    item: "network",
  },
  {
    path: "/secgroups/:id/secrules/:id?",
    component: ModifySecrules,
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
