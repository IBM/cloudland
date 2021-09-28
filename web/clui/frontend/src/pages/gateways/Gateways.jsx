/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Card, Button, Popconfirm, message, Input } from "antd";
import { gwListApi, delGWInfor } from "../../service/gateways";
import DataTable from "../../components/DataTable/DataTable";
import "./gateways.css";
import { withTranslation } from "react-i18next";

const { Search } = Input;

class Gateways extends Component {
  constructor(props) {
    super(props);
    this.state = {
      gateways: [],
      filteredList: [],
      isLoaded: false,
      interfaces: "",
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
      title: this.props.t("Interfaces"),
      dataIndex: "Interfaces",
      width: 140,
      align: "center",
      render: (Interfaces) => (
        <span>
          {Interfaces.map((iface) => {
            return iface.Address.Address + " ";
          })}
        </span>
      ),
    },
    {
      title: this.props.t("Subnets"),
      dataIndex: "Subnets",
      width: 130,
      align: "center",
      render: (Subnets) => (
        <span>
          {Subnets.map((subnet) => {
            return subnet.Gateway + " ";
          })}
        </span>
      ),
    },
    {
      title: this.props.t("Status"),
      dataIndex: "Status",
      align: "center",
    },
    {
      title: "Hyper",
      align: "center",
      className: JSON.parse(sessionStorage.loginInfo).isAdmin
        ? ""
        : "columnHidden",

      render: (record) => (
        <span>
          {record.Hyper},{record.Peer}
        </span>
      ),
    },
    {
      title: this.props.t("Owner"),
      dataIndex: "OwnerInfo.name",
      align: "center",
      className: JSON.parse(sessionStorage.loginInfo).isAdmin
        ? ""
        : "columnHidden",
    },
    {
      title: "Zone",
      dataIndex: "Zone.Name",
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
                this.props.history.push("/gateways/" + record.ID);
              }}
            >
              {t("Edit")}
            </Button>
            <Popconfirm
              title={t("Doyouwanttodelete")}
              okText={t("yes")}
              cancelText={t("no")}
              onConfirm={() => {
                delGWInfor(record.ID).then((res) => {
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

  //it will executed while initting component
  componentDidMount() {
    const _this = this;
    gwListApi()
      .then((res) => {
        _this.setState({
          gateways: res.gateways,
          filteredList: res.gateways,
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
    gwListApi(offset, limit)
      .then((res) => {
        _this.setState({
          gateways: res.gateways,
          filteredList: res.gateways,
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
  createGateways = () => {
    this.props.history.push("/gateways/new");
  };
  filter = (event) => {
    this.getFilteredList(event.target.value);
  };
  getFilteredList = (word) => {
    var keyword = word.toLowerCase();
    if (keyword) {
      this.setState({
        filteredList: this.state.gateways.filter(
          (item) =>
            item.ID.toString().indexOf(keyword) > -1 ||
            item.Name.toLowerCase().indexOf(keyword) > -1 ||
            item.Status.toLowerCase().indexOf(keyword) > -1 ||
            item.Hyper.toString().indexOf(keyword) > -1
        ),
      });
    } else {
      this.setState({
        filteredList: this.state.gateways,
      });
    }
  };
  render() {
    const { t } = this.props;

    return (
      <Card
        title={
          t("Gateway_Manage_Panel") +
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
              onClick={this.createGateways}
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
          onPaginationChange={this.onPaginationChange}
          onShowSizeChange={this.onShowSizeChange}
          pageSizeOptions={this.state.pageSizeOptions}
          loading={!this.state.isLoaded}
        />
      </Card>
    );
  }
}

export default withTranslation()(Gateways);
