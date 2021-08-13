/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/

import request from "../utils/request";
export function flavorsListApi(offset, limit) {
  return request({
    url: "/api/flavors",
    method: "get",
    params: {
      offset,
      limit,
    },
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
