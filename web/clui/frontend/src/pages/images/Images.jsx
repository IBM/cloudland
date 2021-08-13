/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Card, Table, Button, Popconfirm, message } from "antd";
import { imagesListApi, delImgInfor } from "../../api/images";

class Images extends Component {
  constructor(props) {
    super(props);
    this.state = {
      images: [],
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
      dataIndex: "ID",
      key: "ID",
      align: "center",
      width: 70,
    },
    {
      title: "Name",
      dataIndex: "Name",
      align: "center",
    },
    {
      title: "Format",
      dataIndex: "Format",
      align: "center",
    },
    {
      title: "Status",
      dataIndex: "Status",
      align: "center",
    },
    {
      title: "Created At",
      dataIndex: "CreatedAt",
      align: "center",
    },
    {
      title: "OS Version",
      dataIndex: "OsVersion",
      align: "center",
    },
    {
      title: "Hypervisor Type",
      dataIndex: "VirtType",
      align: "center",
    },
    {
      title: "Default Username",
      dataIndex: "UserName",
      align: "center",
    },
    {
      title: "Architecture",
      dataIndex: "Architecture",
      align: "center",
    },
    {
      title: "Action",
      align: "center",
      render: (txt, record, index) => {
        return (
          <div>
            <Popconfirm
              title="确定删除此项?"
              onCancel={() => {
                console.log("用户取消删除");
              }}
              onConfirm={() => {
                console.log("onClick-delete:", record);
                //this.props.history.push("/registrys/new/" + record.ID);
                delImgInfor(record.ID).then((res) => {
                  //const _this = this;
                  message.success(res.Msg);
                  this.loadData(this.state.current, this.state.pageSize);

                  console.log("用户~~", res);
                  console.log("用户~~state", this.state);
                });
              }}
            >
              <Button
                style={{ margin: "0 1rem" }}
                type="danger"
                size="small"
                onClick={() => {
                  console.log("用户", record.ID);
                }}
              >
                Delete
              </Button>
            </Popconfirm>
          </div>
        );
      },
    },
  ];
  componentWillMount() {
    const _this = this;
    const limit = this.state.pageSize;
    imagesListApi(this.state.offset, limit)
      .then((res) => {
        console.log("imagesListApi-total:", res.total);
        _this.setState({
          images: res.images,
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
  createImages = () => {
    this.props.history.push("/images/new");
  };
  loadData = (page, pageSize) => {
    console.log("image-loadData~~", page, pageSize);
    const _this = this;
    const offset = (page - 1) * pageSize;
    const limit = pageSize;
    imagesListApi(offset, limit)
      .then((res) => {
        console.log("loadData", res);

        _this.setState({
          images: res.images,
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
    console.log("image-toSelectchange~limit:", offset, limit);
    imagesListApi(offset, limit)
      .then((res) => {
        console.log("loadData", res);
        _this.setState({
          images: res.images,
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
  render() {
    return (
      <Card
        title="Image Manage Panel"
        extra={
          <Button type="primary" onClick={this.createImages}>
            Create
          </Button>
        }
      >
        <Table
          rowKey="ID"
          columns={this.columns}
          bordered
          dataSource={this.state.images}
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
          loading={!this.state.isLoaded}
        ></Table>
      </Card>
    );
  }
}
export default Images;
