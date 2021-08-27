/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Form, Card, Input, Button, message, InputNumber } from "antd";
import { createFlavorApi } from "../../service/flavors";
const layoutButton = {
  labelCol: { span: 8 },
  wrapperCol: { span: 16 },
};
const layoutForm = {
  labelCol: { span: 6 },
  wrapperCol: { span: 10 },
};
class CreateFlavors extends Component {
  constructor(props) {
    super(props);
    //const { getFieldDecorator } = this.props.form;
    this.state = {
      name: "",
      cpu: 1,
      memory: 1,
      disk: 1,
      swap: 0,
      ephemeral: 0,
    };
    console.log("CreateFlavors~~", props);
  }
  // flavorsChange = (e) => {
  //   console.log("valueChange-e", e);
  //   this.setState({
  //     [e.target.name]: e.target.value,
  //   });
  //   console.log("flavorsChange", this.state);
  // };
  listFlavors = () => {
    this.props.history.push("/flavors");
  };
  handleSubmit = (event) => {
    console.log("handleSubmit-state:", this.state);
    console.log("handleSubmit:", event);
    console.log("handleSubmit-event.target.value:", event.target.value);
    event.preventDefault();
    this.props.form.validateFieldsAndScroll((err) => {
      //   let _this = this;
      if (!err) {
        // this.setState({
        //   flavors: {
        //     Name: this.state.name,
        //     Cpu: parseInt(this.state.cpu),
        //     Memory: parseInt(this.state.memory),
        //     Disk: parseInt(this.state.disk),
        //     Swap: parseInt(this.state.swap),
        //     Ephemeral: parseInt(this.state.ephemeral),
        //   },
        // });
        console.log("提交", this.state.flavors);
        createFlavorApi({
          Name: this.state.name,
          Cpu: this.state.cpu,
          Memory: parseInt(this.state.memory),
          Disk: parseInt(this.state.disk),
          Swap: parseInt(this.state.swap),
          Ephemeral: parseInt(this.state.ephemeral),
        })
          .then((res) => {
            console.log("handleSubmit-res-createFlavorApi:", res);
            this.props.history.push("/flavors");
          })
          .catch((err) => {
            console.log("handleSubmit-error:", err);
          });
      } else {
        message.error(" input wrong information");
      }
    });
  };
  render() {
    return (
      <Card
        title={"Create New Flavor"}
        extra={
          <Button
            style={{ float: "right" }}
            type="primary"
            onClick={this.listFlavors}
          >
            Return
          </Button>
        }
      >
        <Form
          layout="horizontal"
          onSubmit={(e) => this.handleSubmit(e)}
          wrapperCol={{ ...layoutForm.wrapperCol }}
        >
          <Form.Item
            label="Name"
            name="Name"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("Name", {
              rules: [
                {
                  required: true,
                },
              ],
            })(
              <Input
                name="name"
                onChange={(e) => this.setState({ name: e.target.value })}
              />
            )}
          </Form.Item>
          <Form.Item
            label="CPU"
            name="Cpu"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("Cpu", {
              rules: [
                {
                  required: true,
                },
              ],
            })(
              <Input
                name="cpu"
                onChange={(e) => this.setState({ cpu: e.target.value })}
              />
            )}
          </Form.Item>
          <Form.Item
            label="Memory(M)"
            name="Memory"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("Memory", {
              rules: [
                {
                  required: true,
                },
              ],
            })(
              <Input
                name="memory"
                onChange={(e) => this.setState({ memory: e.target.value })}
              />
            )}
          </Form.Item>
          <Form.Item
            label="Disk(G)"
            name="Disk"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("Disk", {
              rules: [
                {
                  required: true,
                },
              ],
            })(
              <Input
                name="disk"
                onChange={(e) => this.setState({ disk: e.target.value })}
              />
            )}
          </Form.Item>
          <Form.Item
            label="Swap(G)"
            name="Swap"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("Swap", {
              rules: [],
              initialValue: 0,
            })(
              <Input
                name="swap"
                onChange={(e) => this.setState({ swap: e.target.value })}
              />
            )}
          </Form.Item>
          <Form.Item
            label="Ephemeral(G)"
            name="Ephemeral"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("Ephemeral", {
              rules: [],
              initialValue: 0,
            })(
              <Input
                name="ephemeral"
                onChange={(e) => this.setState({ ephemeral: e.target.value })}
              />
            )}
          </Form.Item>

          <Form.Item
            wrapperCol={{ ...layoutButton.wrapperCol, offset: 8 }}
            labelCol={{ span: 6 }}
          >
            {
              <Button type="primary" htmlType="submit">
                Create New Flavor
              </Button>
            }
          </Form.Item>
        </Form>
      </Card>
    );
  }
}
export default Form.create({ name: "createFlavors" })(CreateFlavors);
