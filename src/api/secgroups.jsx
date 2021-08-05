import request from "../utils/request";

export function secgroupsListApi() {
  return request({
    url: "/api/secgroups",
    method: "get",
  });
}
