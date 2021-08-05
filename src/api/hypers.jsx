import request from "../utils/request";

export function hypersListApi() {
  return request({
    url: "/api/hypers",
    method: "get",
  });
}
