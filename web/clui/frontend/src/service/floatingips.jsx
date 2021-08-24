/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import request from "../utils/request";

export function floatingipsListApi(paramsObj) {
  return request({
    url: "/api/floatingips",
    method: "get",
    params: paramsObj ? paramsObj : {},
  });
}
export function createFloatingipApi(objFl) {
  return request({
    url: "/api/floatingips/new",
    method: "post",
    data: objFl,
  });
}
export function delFloatingipInfor(flid) {
  return request({
    url: `/api/floatingips/${flid}`,
    method: "delete",
  });
}
