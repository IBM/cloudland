/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import request from "../utils/request";

export function ocpListApi(paramsObj) {
  return request({
    url: "/api/openshifts",
    method: "get",
    params: paramsObj ? paramsObj : {},
  });
}
export function createOcpApi(objOcp) {
  return request({
    url: "/api/openshifts/new",
    method: "post",
    data: objOcp,
  });
}
export function getOcpInforById(ocpid) {
  return request({
    url: `/api/openshifts/${ocpid}`,
    method: "get",
  });
}
export function editOcpInfor(ocpid, obj) {
  return request({
    url: `/api/openshifts/${ocpid}`,
    method: "post",
    data: obj,
  });
}
export function delOcpInfor(ocpid) {
  return request({
    url: `/api/openshifts/${ocpid}`,
    method: "delete",
  });
}
