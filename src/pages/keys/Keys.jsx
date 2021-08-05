import React, { Component } from "react";
import { Card, Table, Button, Popconfirm } from "antd";
import { keysListApi } from "../../api/keys";
const columns = [
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
  },
  {
    title: "Owner",
    dataIndex: "OwnerInfo.name",
  },
  {
    title: "Created At",
    dataIndex: "CreatedAt",
  },
  {
    title: "Action",
    render: (txt, record, index) => {
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
class Keys extends Component {
  constructor(props) {
    super(props);
    this.state = {
      keys: [],
      isLoaded: false,
    };
  }
  componentDidMount() {
    const _this = this;
    keysListApi()
      .then((res) => {
        console.log("componentDidMount-keys:", res);
        _this.setState({
          keys: res.keys,
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
  demo = () => {
    console.log("11");
  };
  render() {
    return (
      <Card
        title="Key Manage Panel"
        extra={
          <Button type="primary" size="small" onClick={this.demo}>
            Create
          </Button>
        }
      >
        <Table
          rowKey="ID"
          columns={columns}
          bordered
          dataSource={this.state.keys}
        ></Table>
      </Card>
    );
  }
}
export default Keys;
