/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import moment from "moment";
import { Card, Button, Popconfirm, message, Input } from "antd";
import { keysListApi, delKeyInfor } from "../../service/keys";
import DataTable from "../../components/DataTable/DataTable";

const { Search } = Input;

class Keys extends Component {
  constructor(props) {
    super(props);
    this.state = {
      keys: [],
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
      title: "ID",
      key: "ID",
      width: 80,
      align: "center",
      dataIndex: "ID",
      //render: (txt, record, index) => index + 1,
    },
    {
      title: "Name",
      dataIndex: "Name",
      align: "center",
    },
    {
      title: "Owner",
      dataIndex: "OwnerInfo.name",
      align: "center",
    },
    {
      title: "Created At",
      dataIndex: "CreatedAt",
      align: "center",
      render: (record) => (
        <span>{moment(record).format("YYYY-MM-DD HH:mm:ss")}</span>
      ),
    },
    {
      title: "Action",
      align: "center",
      render: (txt, record, index) => {
        return (
          <div>
            <Popconfirm
              title="Are you sure to delete?"
              onCancel={() => {
                console.log("cancelled");
              }}
              onConfirm={() => {
                console.log("confirmed");
                delKeyInfor(record.ID).then((res) => {
                  message.success(res.Msg);
                  this.loadData(this.state.current, this.state.pageSize);
                });              }}
            >
              <Button style={{ margin: "0 1rem" }} type="danger" size="small">
                Delete
              </Button>
            </Popconfirm>
          </div>
        );
      },
    },
  ];
  componentDidMount() {
    const _this = this;
    keysListApi()
      .then((res) => {
        console.log("componentDidMount-keys:", res);
        _this.setState({
          keys: res.keys,
          filteredList: res.keys,
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
    console.log("key-loadData~~", page, pageSize);
    const _this = this;
    const offset = (page - 1) * pageSize;
    const limit = pageSize;
    keysListApi(offset, limit)
      .then((res) => {
        console.log("loadData", res);
        _this.setState({
          keys: res.keys,
          filteredList: res.keys,
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
    console.log("key-toSelectchange~limit:", offset, limit);
    keysListApi(offset, limit)
      .then((res) => {
        console.log("loadData", res);
        _this.setState({
          keys: res.keys,
          filteredList: res.keys,
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

  createKey = () => {
    this.props.history.push("/keys/new");
  };

  filter = (event) => {
    console.log("event-filter", event.target.value);
    this.getFilteredList(event.target.value);
  };
  getFilteredList = (word) => {
    console.log("getFilteredListr-keyword-ocp", word);
    var keyword = word.toLowerCase();
    if (keyword) {
      this.setState({
        filteredList: this.state.keys.filter(
          (item) =>
            item.ID.toString().indexOf(keyword) > -1 ||
            item.Name.toLowerCase().indexOf(keyword) > -1
        ),
      });

      console.log("filteredList", this.state.filteredList);
    } else {
      this.setState({
        filteredList: this.state.keys,
      });
    }
  };
  render() {
    return (
      <Card
        title={
          "Key Manage Panel" + "(Total: " + this.state.filteredList.length + ")"
        }
        extra={
          <div>
            <Search
              placeholder="Search..."
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
              onClick={this.createKey}
            >
              Create
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

export default Keys;
