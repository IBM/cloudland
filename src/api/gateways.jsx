import request from "../utils/request";

export function gatewaysListApi() {
  return request({
    url: "/api/gateways",
    method: "get",
  });
}
