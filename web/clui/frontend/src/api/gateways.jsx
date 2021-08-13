/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import request from "../utils/request";

export function gatewaysListApi() {
  return request({
    url: "/api/gateways",
    method: "get",
  });
}
