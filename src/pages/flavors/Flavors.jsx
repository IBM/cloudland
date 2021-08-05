import React, { Component } from "react";
import { Card, Table, Button, Popconfirm } from "antd";
import { flavorsListApi } from "../../api/flavors";
const columns = [
  {
    title: "ID",
    dataIndex: "ID",
    key: "ID",
    width: 80,
    align: "center",
  },
  {
    title: "Name",
    dataIndex: "Name",
  },
  {
    title: "CPU",
    dataIndex: "Cpu",
  },
  {
    title: "Memory",
    dataIndex: "Memory",
  },
  {
    title: "Disk",
    dataIndex: "Disk",
  },
  {
    title: "Swap",
    dataIndex: "Swap",
  },
  {
    title: "Ephemeral",
    dataIndex: "Ephemeral",
  },
  {
    title: "Action",
    render: (txt, record, index) => {
      return (
        <div>
          <Button type="primary" size="small">
            Edit
          </Button>
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
class Flavors extends Component {
  constructor(props) {
    super(props);
    this.state = {
      flavors: [],
      isLoaded: false,
    };
  }
  componentDidMount() {
    const _this = this;
    flavorsListApi()
      .then((res) => {
        _this.setState({
          flavors: res.flavors,
          isLoaded: true,
        });
        console.log("flavors:", res);
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
        title="Flavor Manage Panel"
        extra={<Button type="primary">Create</Button>}
      >
        <Table
          rowKey="ID"
          columns={columns}
          dataSource={this.state.flavors}
        ></Table>
      </Card>
    );
  }
}
export default Flavors;
