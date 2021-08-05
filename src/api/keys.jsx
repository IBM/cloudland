import request from "../utils/request";

export function keysListApi() {
  return request({
    url: "/api/keys",
    method: "get",
  });
}
