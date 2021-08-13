/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import request from "../utils/request";

export function keysListApi() {
  return request({
    url: "/api/keys",
    method: "get",
  });
}
