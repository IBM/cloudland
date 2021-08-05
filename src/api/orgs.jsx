import request from "../utils/request";

export function orgsListApi() {
  return request({
    url: "/api/orgs",
    method: "get",
  });
}
