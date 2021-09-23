/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { withTranslation } from "react-i18next";

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
      title: this.props.t("HyperID"),
      dataIndex: "ID",
      width: 80,
      align: "center",
      //render: (txt, record, index) => index + 1,
    },
    {
      title: this.props.t("Hostname"),
      dataIndex: "Hostname",
      align: "center",
    },
    {
      title: this.props.t("ParentID"),
      dataIndex: "Parentid",
      align: "center",
    },
    {
      title: this.props.t("Children"),
      dataIndex: "Children",
      align: "center",
    },
    {
      title: this.props.t("HostIP"),
      dataIndex: "HostIP",
      align: "center",
    },
    {
      title: this.props.t("Status"),
      dataIndex: "Status",
      align: "center",
    },
    {
      title: this.props.t("Zone"),
      dataIndex: "Zone.Name",
      align: "center",
    },
    {
      title: this.props.t("Cpu"),
      dataIndex: "Resource.Cpu",
      align: "center",
    },
    {
      title: this.props.t("Memory") + "(K)",
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
      title: this.props.t("Disk") + "(B)",
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
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
  }
  loadData = (page, pageSize) => {
    const _this = this;
    const offset = (page - 1) * pageSize;
    const limit = pageSize;
    hypersListApi(offset, limit)
      .then((res) => {
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
  toSelectchange = (page, num) => {
    const offset = (page - 1) * num;
    const limit = num;
    this.loadData(offset, limit);
  };
  onPaginationChange = (e) => {
    this.loadData(e, this.state.pageSize);
  };
  onShowSizeChange = (current, pageSize) => {
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
export default withTranslation()(Hypers);
