/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/

import request from "../utils/request";
export function flavorsListApi(paramsObj) {
  return request({
    url: "/api/flavors",
    method: "get",
    params: paramsObj ? paramsObj : {},
  });
}
export function createFlavorApi(objFla) {
  return request({
    url: "/api/flavors/new",
    method: "post",
    data: objFla,
  });
}
export function delFlavorInfor(flavorid) {
  return request({
    url: `/api/flavors/${flavorid}`,
    method: "delete",
  });
}
