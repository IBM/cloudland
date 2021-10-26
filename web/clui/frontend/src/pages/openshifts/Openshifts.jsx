/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Card, Button, Popconfirm, Input, message } from "antd";
import { withTranslation } from "react-i18next";

import { ocpListApi, delOcpInfor } from "../../service/openshifts";
import DataTable from "../../components/DataTable/DataTable";
const { Search } = Input;
class Openshifts extends Component {
  constructor(props) {
    super(props);
    this.state = {
      openshifts: [],
      filteredList: [],
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
      title: this.props.t("ID"),
      dataIndex: "ID",
      key: "ID",
      width: 80,
      align: "center",
    },
    {
      title: this.props.t("Cluster_Name"),
      dataIndex: "ClusterName",
      align: "center",
    },
    {
      title: this.props.t("Base_Domain"),
      dataIndex: "BaseDomain",
      align: "center",
    },
    {
      title: this.props.t("Console"),
      dataIndex: "console",
      align: "center",
    },
    {
      title: this.props.t("Version"),
      dataIndex: "Version",
      align: "center",
    },
    {
      title: this.props.t("Flavors"),
      dataIndex: "Flavor",
      align: "center",
    },
    {
      title: this.props.t("HA"),
      dataIndex: "Haflag",
      // width: 50,
      align: "center",
    },
    {
      title: this.props.t("N-Workers"),
      dataIndex: "WorkerNum",
      align: "center",
    },
    {
      title: this.props.t("Status"),
      dataIndex: "Status",
      align: "center",
    },
    {
      title: this.props.t("Action"),
      align: "center",
      render: (txt, record, index) => {
        const { t } = this.props;
        return (
          <div>
            <Button
              style={{
                marginTop: "10px",
              }}
              type="primary"
              size="small"
              onClick={() => {
                this.props.history.push("/openshifts/" + record.ID);
              }}
            >
              {t("Edit")}
            </Button>
            <Popconfirm
              title={t("Doyouwanttodelete")}
              okText={t("yes")}
              cancelText={t("no")}
              onConfirm={() => {
                delOcpInfor(record.ID).then((res) => {
                  message.success(res.Msg);
                  this.loadData(this.state.current, this.state.pageSize);
                });
              }}
            >
              <Button
                style={{
                  margin: "5px",
                  marginRight: "0px",
                  marginTop: "10px",
                }}
                type="danger"
                size="small"
              >
                {t("Delete")}
              </Button>
            </Popconfirm>
          </div>
        );
      },
    },
  ];
  componentDidMount() {
    const _this = this;
    ocpListApi()
      .then((res) => {
        _this.setState({
          openshifts: res.openshifts,
          filteredList: res.openshifts,
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
  createOpenshift = () => {
    this.props.history.push("/openshifts/new");
  };
  //it's to load data to get Ocp data
  loadData = (page, pageSize) => {
    const _this = this;
    const offset = (page - 1) * pageSize;
    const limit = pageSize;
    ocpListApi(offset, limit)
      .then((res) => {
        _this.setState({
          openshifts: res.openshifts,
          filteredList: res.openshifts,
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
  //go to selected page while clicking
  toSelectchange = (page, num) => {
    const offset = (page - 1) * num;
    const limit = num;
    this.loadData(offset, limit);
  };
  //pageSize changed
  onPaginationChange = (e) => {
    this.loadData(e, this.state.pageSize);
  };
  onShowSizeChange = (current, pageSize) => {
    this.toSelectchange(current, pageSize);
  };
  //show the filtered results while input keyword
  filter = (event) => {
    this.getFilteredList(event.target.value);
  };
  getFilteredList = (word) => {
    var keyword = word.toLowerCase();
    if (keyword) {
      this.setState({
        filteredList: this.state.openshifts.filter(
          (item) =>
            item.ID.toString().indexOf(keyword) > -1 ||
            item.BaseDomain.toLowerCase().indexOf(keyword) > -1 ||
            item.ClusterName.toLowerCase().indexOf(keyword) > -1 ||
            item.Flavor.toString().indexOf(keyword) > -1 ||
            item.Haflag.toLowerCase().indexOf(keyword) > -1 ||
            item.Status.toLowerCase().indexOf(keyword) > -1 ||
            item.WorkerNum.toString().indexOf(keyword) > -1 ||
            item.Version.toLowerCase().indexOf(keyword) > -1
        ),
      });
    } else {
      this.setState({
        filteredList: this.state.openshifts,
      });
    }
  };
  render() {
    const { t } = this.props;
    return (
      <Card
        title={
          t("Openshift_Cluster_Manage_Panel") +
          "(" +
          t("Total") +
          ":" +
          this.state.filteredList.length +
          ")"
        }
        extra={
          <div>
            <Search
              placeholder={t("Search_placeholder")}
              onChange={this.filter}
              enterButton
            />
            <Button
              style={{
                float: "right",
                paddingLeft: "10px",
                paddingLight: "10px",
              }}
              type="primary"
              onClick={this.createOpenshift}
            >
              {t("Create")}
            </Button>
          </div>
        }
      >
        <DataTable
          rowKey="ID"
          columns={this.columns}
          dataSource={this.state.filteredList}
          bordered
          total={this.state.filteredList.length}
          pageSize={this.state.pageSize}
          // scroll={{ y: 600, x: 600 }}
          onPaginationChange={this.onPaginationChange}
          onShowSizeChange={this.onShowSizeChange}
          pageSizeOptions={this.state.pageSizeOptions}
          loading={!this.state.isLoaded}
        />
      </Card>
    );
  }
}

export default withTranslation()(Openshifts);
