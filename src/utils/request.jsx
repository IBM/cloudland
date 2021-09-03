/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import { message } from "antd";
import axios from "axios";
import { getToken } from "./auth";

const instance = axios.create({
  baseURL: "https://cloudland.pic.cdl.ibm.com",
  timeout: 5000,
  headers: {
    "Content-Type": "application/json;charset=UTF-8",
    "Allow-Control-Allow-Origin": "*",
  },
});
//全局请求拦截，发送请求之前执行
instance.interceptors.request.use(
  (config) => {
    console.log("getToken():", getToken());

    if (getToken()) {
      config.headers.common["X-Auth-Token"] = getToken();
    } else {
      delete config.headers.common["x-auth-token"];
      // router.push({
      //   name: "login",
      // });
    }

    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);
//请求返回之后执行
instance.interceptors.response.use(
  (response) => {
    console.log("axios-response", response);
    return response.data;
  },
  (error) => {
    if (error.response) {
      switch (error.response.status) {
        case 401:
          console.log("Need Login");
          window.location = "/login";
          break;
        default:
          message.error(error.response.data.ErrorMsg, 5);
          break;
      }
    }
    console.log("axios-error", error.response);
    return Promise.reject(error);
    // console.log("axios-error", error.response);
    // return Promise.reject(error);
  }
);
export default instance;
