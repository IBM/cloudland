import request from "../utils/request";

export function secrulesListApi() {
  return request({
    url: "/api/secrules",
    method: "get",
  });
}
