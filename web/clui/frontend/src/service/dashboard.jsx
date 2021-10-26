/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/

import request from "../utils/request";
export function getResourceData(paramsObj) {
  return request({
    url: "/api/dashboard/getdata",
    method: "get",
    params: paramsObj ? paramsObj : {},
  });
}
