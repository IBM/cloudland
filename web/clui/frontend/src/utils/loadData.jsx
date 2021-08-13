/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React from "react";
import { insListApi } from "../../api/instances";
export default function loadData(page, pageSize) {
  console.log("ins-loadData~~", page, pageSize);
  const _this = this;
  const offset = (page - 1) * pageSize;
  const limit = pageSize;
  insListApi(offset, limit)
    .then((res) => {
      console.log("loadData", res);
      _this.setState({
        instances: res.instances,
        isLoaded: true,
        total: res.total,
        pageSize: limit,
        current: page,
      });
      console.log("loadData-page-", page, _this.state);
    })
    .catch((error) => {
      _this.setState({
        isLoaded: false,
        error: error,
      });
    });
}
