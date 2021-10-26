/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { withTranslation } from "react-i18next";

import { Card, Button, Popconfirm, message, Input } from "antd";
import {
  floatingipsListApi,
  delFloatingipInfor,
} from "../../service/floatingips";
import DataTable from "../../components/DataTable/DataTable";
const { Search } = Input;
class Floatingips extends Component {
  constructor(props) {
    super(props);
    this.state = {
      floatingips: [],
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
      title: this.props.t("FloatingIps"),
      dataIndex: "FipAddress",
      align: "center",
    },
    {
      title: this.props.t("InternalIP"),
      dataIndex: "IntAddress",
      align: "center",
    },
    {
      title: this.props.t("Instance"),
      dataIndex: "Instance.Hostname",
      align: "center",
    },
    {
      title: this.props.t("Zone"),
      dataIndex: "Instance.Zone.Name",
      align: "center",
    },
    {
      title: this.props.t("Action"),
      align: "center",
      render: (txt, record, index) => {
        const { t } = this.props;
        return (
          <div>
            <Popconfirm
              title={t("Doyouwanttodelete")}
              okText={t("yes")}
              cancelText={t("no")}
              onConfirm={() => {
                delFloatingipInfor(record.ID).then((res) => {
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
  createFloatingips = () => {
    this.props.history.push("/floatingips/new");
  };
  //it will executed while initting component
  componentDidMount() {
    const _this = this;
    floatingipsListApi()
      .then((res) => {
        _this.setState({
          floatingips: res.floatingips,
          filteredList: res.floatingips,
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
  //loading data while refreshing
  loadData = (page, pageSize) => {
    const _this = this;
    const offset = (page - 1) * pageSize;
    const limit = pageSize;
    floatingipsListApi(offset, limit)
      .then((res) => {
        _this.setState({
          floatingips: res.floatingips,
          filteredList: res.floatingips,
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
        filteredList: this.state.floatingips.filter(
          (item) =>
            item.ID.toString().indexOf(keyword) > -1 ||
            item.FipAddress.indexOf(keyword) > -1 ||
            item.IntAddress.indexOf(keyword) > -1 ||
            item.Instance.Hostname.toLowerCase().indexOf(keyword) > -1
        ),
      });
    } else {
      this.setState({
        filteredList: this.state.floatingips,
      });
    }
  };
  render() {
    const { t } = this.props;
    return (
      <Card
        title={
          t("Floating_IP_Manage_Panel") +
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
                paddingRight: "10px",
              }}
              type="primary"
              onClick={this.createFloatingips}
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
export default withTranslation()(Floatingips);
