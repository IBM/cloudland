import React, { Component } from "react";
import { Card, Table, Button, Popconfirm } from "antd";
import { imagesListApi } from "../../api/images";
const columns = [
  {
    title: "ID",
    dataIndex: "ID",
    key: "ID",
  },
  {
    title: "Name",
    dataIndex: "Name",
  },
  {
    title: "Format",
    dataIndex: "Format",
  },
  {
    title: "Status",
    dataIndex: "Status",
  },
  {
    title: "Created At",
    dataIndex: "CreatedAt",
  },
  {
    title: "OS Version",
    dataIndex: "OsVersion",
  },
  {
    title: "Hypervisor Type",
    dataIndex: "VirtType",
  },
  {
    title: "Default Username",
    dataIndex: "UserName",
  },
  {
    title: "Architecture",
    dataIndex: "Architecture",
  },
  {
    title: "Action",
    render: () => {
      return (
        <div>
          <Popconfirm
            title="确定删除此项?"
            onCancel={() => {
              console.log("用户取消删除");
            }}
            onConfirm={() => {
              console.log("用户确认删除");
              //此处调用api接口进行相关操作
            }}
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
class Images extends Component {
  constructor(props) {
    super(props);
    this.state = {
      images: [],
      isLoaded: false,
    };
  }
  componentDidMount() {
    const _this = this;
    imagesListApi()
      .then((res) => {
        _this.setState({
          images: res.images,
          isLoaded: true,
        });
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
  }
  render() {
    return (
      <Card
        title="Image Manage Panel"
        extra={<Button type="primary">Create</Button>}
      >
        <Table
          rowKey="ID"
          columns={columns}
          dataSource={this.state.images}
        ></Table>
      </Card>
    );
  }
}
export default Images;
