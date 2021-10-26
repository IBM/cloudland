/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { withTranslation } from "react-i18next";

import { Card, Button, Popconfirm, message, Input } from "antd";
import { subnetsListApi, delSubInfor } from "../../service/subnets";

import DataTable from "../../components/DataTable/DataTable";

const { Search } = Input;
class Subnets extends Component {
  constructor(props) {
    super(props);
    this.state = {
      subnets: [],
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
      title: this.props.t("Name"),
      dataIndex: "Name",
      align: "center",
    },
    {
      title: this.props.t("Network"),
      dataIndex: "Network",
      align: "center",
    },
    {
      title: this.props.t("Netmask"),
      dataIndex: "Netmask",
      align: "center",
    },
    {
      title: this.props.t("Zone"),
      dataIndex: "Zones",
      align: "center",
      render: (Zones) => (
        <span>
          {Zones.map((zones) => {
            return zones.Name;
          })}
        </span>
      ),
    },
    {
      title: this.props.t("Vlan"),
      dataIndex: "Vlan",
      align: "center",
    },
    {
      title: this.props.t("Hyper"),
      dataIndex: "Netlink.Hyper",
      align: "center",
    },
    {
      title: this.props.t("Owner"),
      dataIndex: "OwnerInfo.name",
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
                this.props.history.push("/subnets/" + record.ID);
              }}
            >
              {t("Edit")}
            </Button>
            <Popconfirm
              title={t("Doyouwanttodelete")}
              okText={t("yes")}
              cancelText={t("no")}
              onCancel={() => {
                this.props.history.push("/subnets");
              }}
              onConfirm={() => {
                delSubInfor(record.ID)
                  .then((res) => {
                    message.success(res.Msg);
                    this.loadData(this.state.current, this.state.pageSize);
                  })
                  .catch((err) => {
                    console.log("subnet-err", err);
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
    subnetsListApi()
      .then((res) => {
        _this.setState({
          subnets: res.subnets,
          filteredList: res.subnets,
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
  //it's to load data to get subnet data
  loadData = (page, pageSize) => {
    const _this = this;
    const offset = (page - 1) * pageSize;
    const limit = pageSize;
    subnetsListApi(offset, limit)
      .then((res) => {
        _this.setState({
          subnets: res.subnets,
          filteredList: res.subnets,
          isLoaded: true,
          total: res.total,
          pageSize: limit,
          current: page,
        });
      })
      .catch((error) => {
        message.error(error.response.data.ErrorMsg);
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
  onPaginationChange = (e) => {
    this.loadData(e, this.state.pageSize);
  };
  onShowSizeChange = (current, pageSize) => {
    this.toSelectchange(current, pageSize);
  };
  createSubnets = () => {
    this.props.history.push("/subnets/new");
  };
  //show the filtered results while input keyword
  filter = (event) => {
    this.getFilteredList(event.target.value);
  };
  getFilteredList = (word) => {
    var keyword = word.toLowerCase();
    if (keyword) {
      this.setState({
        filteredList: this.state.subnets.filter(
          (item) =>
            item.ID.toString().indexOf(keyword) > -1 ||
            item.Name.toLowerCase().indexOf(keyword) > -1 ||
            item.Network.toLowerCase().indexOf(keyword) > -1 ||
            item.Network.toLowerCase().indexOf(keyword) > -1 ||
            item.Vlan.toString().indexOf(keyword) > -1
        ),
      });
    } else {
      this.setState({
        filteredList: this.state.subnets,
      });
    }
  };
  render() {
    const { t } = this.props;
    return (
      <Card
        title={
          t("Subnet_Manage_Panel") +
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
              onClick={this.createSubnets}
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

export default withTranslation()(Subnets);
