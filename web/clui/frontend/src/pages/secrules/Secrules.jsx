/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Card, Button, Popconfirm, message, Input } from "antd";
import { withTranslation } from "react-i18next";

import { secrulesListApi, delSecruleInfor } from "../../service/secrules";
import DataTable from "../../components/DataTable/DataTable";

import "./secrules.css";
const { Search } = Input;

class Secrules extends Component {
  constructor(props) {
    super(props);
    this.state = {
      secrules: [],
      filteredList: [],
      sgID: "",
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
      // render: (href) => <a>{href}</a>,
    },
    {
      title: this.props.t("SecurityGroup"),
      dataIndex: "Secgroup",
      align: "center",
      // render: (href) => <a>{href}</a>,
    },
    {
      title: this.props.t("RemoteIp"),
      dataIndex: "RemoteIp",
      align: "center",
    },
    {
      title: this.props.t("Direction"),
      dataIndex: "Direction",
      align: "center",
    },
    {
      title: this.props.t("Protocol"),
      dataIndex: "Protocol",
      align: "center",
    },
    {
      title: this.props.t("PortMin_Type"),
      dataIndex: "PortMin",
      align: "center",
    },
    {
      title: this.props.t("PortMax_Code"),
      dataIndex: "PortMax",
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
                this.props.history.push(
                  `/secgroups/${record.Secgroup}/secrules/` + record.ID
                );
              }}
            >
              {t("Edit")}
            </Button>
            <Popconfirm
              title={t("Doyouwanttodelete")}
              okText={t("yes")}
              cancelText={t("no")}
              onConfirm={() => {
                delSecruleInfor(record.Secgroup, record.ID).then((res) => {
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
    secrulesListApi(this.props.match.params.id)
      .then((res) => {
        _this.setState({
          secrules: res.secrules,
          filteredList: res.secrules,
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
  createSecrules = () => {
    this.props.history.push(
      `/secgroups/${this.props.match.params.id}/secrules/new`
    );
  };
  listSecgroups = () => {
    this.props.history.push("/secgroups");
  };
  loadData = (page, pageSize) => {
    const _this = this;
    const offset = (page - 1) * pageSize;
    const limit = pageSize;
    const sgID = this.props.match.params.id;
    secrulesListApi(sgID, { offset, limit })
      .then((res) => {
        _this.setState({
          secrules: res.secrules,
          filteredList: res.secrules,
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
    const sgID = this.props.match.params.id;

    this.loadData(sgID, { offset, limit });
  };
  onPaginationChange = (e) => {
    this.loadData(e, this.state.pageSize);
  };
  onShowSizeChange = (current, pageSize) => {
    this.toSelectchange(current, pageSize);
  };
  filter = (event) => {
    this.getFilteredList(event.target.value);
  };
  getFilteredList = (word) => {
    var keyword = word.toLowerCase();
    if (keyword) {
      this.setState({
        filteredList: this.state.secrules.filter(
          (item) =>
            item.ID.toString().indexOf(keyword) > -1 ||
            item.Protocol.toLowerCase().indexOf(keyword) > -1 ||
            item.Direction.toLowerCase().indexOf(keyword) > -1 ||
            item.RemoteIp.indexOf(keyword) > -1
        ),
      });
    } else {
      this.setState({
        filteredList: this.state.secrules,
      });
    }
  };
  render() {
    const { t } = this.props;
    return (
      <Card
        title={
          t("Security_Rules_Manage_Panel") +
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
                marginLeft: "50px",
                paddingLeft: "10px",
                paddingRight: "10px",
              }}
              type="primary"
              onClick={this.createSecrules}
            >
              {t("Create")}
            </Button>
            <Button
              style={{
                float: "right",
                marginLeft: "10px",
                paddingLeft: "10px",
                paddingRight: "10px",
              }}
              type="primary"
              // size="small"
              onClick={this.listSecgroups}
            >
              {t("Return")}
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

export default withTranslation()(Secrules);
