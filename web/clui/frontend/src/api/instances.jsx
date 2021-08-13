/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import request from "../utils/request";
// export function insListApi() {
//   return request({
//     url: "/api/instances",
//     method: "get",
//   });
// }
export function insListApi(paramsObj) {
  return request({
    url: "/api/instances/",
    method: "get",
    params: paramsObj ? paramsObj : {},
  });
}
export function createInsApi(objInstance) {
  return request({
    url: "/api/instances/new",
    method: "post",
    data: objInstance,
  });
}
export function getInsInforById(instanceid) {
  return request({
    url: `/api/instances/${instanceid}`,
    method: "get",
  });
}
export function editInsInfor(instanceid, obj) {
  return request({
    url: `/api/instances/${instanceid}`,
    method: "post",
    data: obj,
  });
}
export function delInsInfor(insid) {
  return request({
    url: `/api/instances/${insid}`,
    method: "delete",
  });
}
