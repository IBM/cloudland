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
  function (config) {
    console.log("getToken():", getToken());
    if (getToken()) {
      config.headers.common["X-Auth-Token"] = getToken();
    } else {
      delete config.headers.common["x-auth-token"];
    }

    return config;
  },
  function (error) {
    return Promise.reject(error);
  }
);
//请求返回之后执行
instance.interceptors.response.use(
  function (response) {
    return response.data;
  },
  function (error) {
    return Promise.reject(error);
  }
);
export default instance;
