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
export function createKeyApi(objKey) {
  return request({
    url: "/api/keys/new",
    method: "post",
    data: objKey,
  });
}
