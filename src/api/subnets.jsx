import request from "../utils/request";

export function subnetsListApi() {
  return request({
    url: "/api/subnets",
    method: "get",
  });
}
