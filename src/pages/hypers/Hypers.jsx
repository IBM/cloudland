/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Card } from "antd";
import { hypersListApi } from "../../service/hypers";
import DataTable from "../../components/DataTable/DataTable";

class Hypers extends Component {
  constructor(props) {
    super(props);
    this.state = {
      hypers: [],
      isLoaded: false,
      total: 0,
      pageSize: 10,
      offset: 0,
      pageSizeOptions: ["5", "10", "15", "20"],
      current: 1,
    };
  }
  columns = [
    {
      title: "HyperID",
      dataIndex: "ID",
      width: 80,
      align: "center",
      //render: (txt, record, index) => index + 1,
    },
    {
      title: "Hostname",
      dataIndex: "Hostname",
      align: "center",
    },
    {
      title: "ParentID",
      dataIndex: "Parentid",
      align: "center",
    },
    {
      title: "Children",
      dataIndex: "Children",
      align: "center",
    },
    {
      title: "HostIP",
      dataIndex: "HostIP",
      align: "center",
    },
    {
      title: "Status",
      dataIndex: "Status",
      align: "center",
    },
    {
      title: "Zone",
      dataIndex: "Zone.Name",
      align: "center",
    },
    {
      title: "CPU",
      dataIndex: "Resource.Cpu",
      align: "center",
    },
    {
      title: "Memory(K)",
      dataIndex: "Resource.Memory",
      align: "center",
      render: (text, record, index) => {
        return (
          <span>
            {record.Resource.Memory}/{record.Resource.MemoryTotal}
          </span>
        );
      },
    },
    {
      title: "Disk(B)",
      dataIndex: "Resource.Disk",
      align: "center",
      render: (text, record, index) => {
        return (
          <span>
            {record.Resource.Disk}/{record.Resource.DiskTotal}
          </span>
        );
      },
    },
  ];
  componentDidMount() {
    const _this = this;
    hypersListApi()
      .then((res) => {
        _this.setState({
          hypers: res.hypers,
          isLoaded: true,
          total: res.total,
        });
        console.log("hyper-hypersListApi:", res);
        console.log("hyper-hypersListApi:", _this.state.hypers);
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
  }
  loadData = (page, pageSize) => {
    console.log("hyper-loadData~~", page, pageSize);
    const _this = this;
    const offset = (page - 1) * pageSize;
    const limit = pageSize;
    hypersListApi(offset, limit)
      .then((res) => {
        console.log("loadData", res);
        _this.setState({
          hypers: res.hypers,
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
  };
  toSelectchange = (page, num) => {
    console.log("toSelectchange", page, num);
    const _this = this;
    const offset = (page - 1) * num;
    const limit = num;
    console.log("hypers-toSelectchange~limit:", offset, limit);
    hypersListApi(offset, limit)
      .then((res) => {
        console.log("loadData", res);
        _this.setState({
          hypers: res.hypers,
          isLoaded: true,
          total: res.total,
          pageSize: limit,
          current: page,
        });
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
  };
  onPaginationChange = (e) => {
    console.log("onPaginationChange", e);
    this.loadData(e, this.state.pageSize);
  };
  onShowSizeChange = (current, pageSize) => {
    console.log("onShowSizeChange:", current, pageSize);
    //当几条一页的值改变后调用函数，current：改变显示条数时当前数据所在页；pageSize:改变后的一页显示条数
    this.toSelectchange(current, pageSize);
  };
  render() {
    return (
      <Card
        title={"Hypervisors View Panel" + "(Total: " + this.state.total + ")"}
      >
        <DataTable
          rowKey="ID"
          columns={this.columns}
          dataSource={this.state.hypers}
          bordered
          total={this.state.total}
          pageSize={this.state.pageSize}
          scroll={{ y: 600 }}
          onPaginationChange={this.onPaginationChange}
          onShowSizeChange={this.onShowSizeChange}
          pageSizeOptions={this.state.pageSizeOptions}
          loading={!this.state.isLoaded}
        />
      </Card>
    );
  }
}
export default Hypers;
