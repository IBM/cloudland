/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import request from "../utils/request";

export function floatingipsListApi() {
  return request({
    url: "/api/floatingips",
    method: "get",
  });
}
