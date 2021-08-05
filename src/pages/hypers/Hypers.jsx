import React, { Component } from "react";
import { Card, Table, Button } from "antd";
import { hypersListApi } from "../../api/hypers";
const columns = [
  {
    title: "HyperID",
    dataIndex: "ID",
    width: 80,
    align: "center",
    //render: (txt, record, index) => index + 1,
  },
  {
    title: "Hostname",
    dataIndex: "Hostname",
  },
  {
    title: "ParentID",
    dataIndex: "Parentid",
  },
  {
    title: "Children",
    dataIndex: "Children",
  },
  {
    title: "HostIP",
    dataIndex: "HostIP",
  },
  {
    title: "Status",
    dataIndex: "Status",
  },
  {
    title: "Zone",
    dataIndex: "Zone.Name",
  },
  {
    title: "CPU",
    dataIndex: "Resource.Cpu",
  },
  {
    title: "Memory(K)",
    dataIndex: "Resource.Memory",
    render: (text, record, index) => {
      return (
        <span>
          {record.Resource.Memory}/{record.Resource.MemoryTotal}
        </span>
      );
    },
  },
  {
    title: "Disk(B)",
    dataIndex: "Resource.Disk",
    render: (text, record, index) => {
      return (
        <span>
          {record.Resource.Disk}/{record.Resource.DiskTotal}
        </span>
      );
    },
  },
];
class Hypers extends Component {
  constructor(props) {
    super(props);
    this.state = {
      hypers: [],
      isLoaded: false,
    };
  }
  componentDidMount() {
    const _this = this;
    hypersListApi()
      .then((res) => {
        _this.setState({
          hypers: res.hypers,
          isLoaded: true,
        });
        console.log("hyper-hypersListApi:", res);
        console.log("hyper-hypersListApi:", _this.state.hypers);
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
        title="Hypervisors View Panel"
        extra={
          <Button
            type="primary"
            size="small"
            //onClick={() => this.props.history.push("`/orgs/${orgsid}`")}
          >
            Create
          </Button>
        }
      >
        <Table
          rowKey="ID"
          columns={columns}
          bordered
          dataSource={this.state.hypers}
        ></Table>
      </Card>
    );
  }
}
export default Hypers;
