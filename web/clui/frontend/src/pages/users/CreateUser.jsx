/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Form, Card, Input, Button, message } from "antd";
import { createUserApi } from "../../service/users";
import "./users.css";
const layoutButton = {
  labelCol: { span: 8 },
  wrapperCol: { span: 16 },
};
const layoutForm = {
  labelCol: { span: 6 },
  wrapperCol: { span: 10 },
  LayoutType: "horizontal",
};
class CreateUser extends Component {
  constructor(props) {
    super(props);
    console.log("CreateUser~~", props);
    this.state = {
      currentData: [],
    };
  }

  listUsers = () => {
    this.props.history.push("/users");
  };
  handleSubmit = (e) => {
    console.log("handleSubmit:", e);
    e.preventDefault();
    this.props.form.validateFieldsAndScroll((err, values) => {
      if (!err) {
        console.log("handleSubmit-value:", values);
        console.log("提交");

        createUserApi(values)
          .then((res) => {
            console.log("handleSubmit-res-createUserApi:", res);
            this.props.history.push("/users");
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
        title={"Create User"}
        extra={
          <Button
            style={{ float: "right" }}
            type="primary"
            onClick={this.listUsers}
          >
            Return
          </Button>
        }
      >
        <Form
          layout={{ ...layoutForm.LayoutType }}
          wrapperCol={{ ...layoutForm.wrapperCol }}
          onSubmit={(e) => this.handleSubmit(e)}
        >
          <Form.Item
            label="User Name"
            name="username"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("username", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.Username,
            })(<Input />)}
          </Form.Item>
          <Form.Item
            name="password"
            label="Password"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("password", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.Password,
            })(<Input type="password" />)}
          </Form.Item>
          <Form.Item
            name="confirm"
            label="Confirm"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("confirm", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.Confirm,
            })(<Input type="password" />)}
          </Form.Item>
          <Form.Item
            wrapperCol={{ ...layoutButton.wrapperCol, offset: 8 }}
            labelCol={{ span: 6 }}
          >
            <Button type="primary" htmlType="submit">
              Create User
            </Button>
          </Form.Item>
        </Form>
      </Card>
    );
  }
}
export default Form.create({ name: "CreateUser" })(CreateUser);
