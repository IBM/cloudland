/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import request from "../utils/request";

export function ocpListApi() {
  return request({
    url: "/api/openshifts",
    method: "get",
  });
}
