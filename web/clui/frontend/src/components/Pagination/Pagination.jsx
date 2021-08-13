/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";

class Pagination extends Component {
    constructor(props) {
        super(props)
        this.state = {
            total: 0,
      pageSize: 10,
      offset: 0,
      pageSizeOptions: ["5", "10", "15", "20"],
      current: 1,

        }
    }

    pagination={{
        //pagination
        total: this.state.total, //total count
        defaultPageSize: this.state.pageSize, //default pageSize
        showSizeChanger: true, //是否显示可以设置几条一页的选项
        onShowSizeChange: (current, pageSize) => {
          console.log("onShowSizeChange:", current, pageSize);
          //当几条一页的值改变后调用函数，current：改变显示条数时当前数据所在页；pageSize:改变后的一页显示条数
          this.toSelectchange(current, pageSize);
        },

        onChange: (current) => {
          this.loadData(current, this.state.pageSize);
        },
        showTotal: () => {
          return "Total " + this.state.total + " items";
        },
        pageSizeOptions: this.state.pageSizeOptions,
      }}
  render() {
    return <div></div>;
  }
}
export default Pagination;
