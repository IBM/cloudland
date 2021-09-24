/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Form, Card, Input, Button, message } from "antd";
import { createUserApi } from "../../service/users";
import { withTranslation } from "react-i18next";
import { compose } from "redux";

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
    const { t } = this.props;
    return (
      <Card
        title={t("Create New User")}
        extra={
          <Button
            style={{ float: "right" }}
            type="primary"
            onClick={this.listUsers}
          >
            {t("Return")}
          </Button>
        }
      >
        <Form
          layout={{ ...layoutForm.LayoutType }}
          wrapperCol={{ ...layoutForm.wrapperCol }}
          onSubmit={(e) => this.handleSubmit(e)}
        >
          <Form.Item
            label={t("Username")}
            name="username"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("username", {
              rules: [
                {
                  required: true,
                },
              ],
            })(<Input />)}
          </Form.Item>
          <Form.Item
            name="password"
            label={t("Password")}
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("password", {
              rules: [
                {
                  required: true,
                },
              ],
            })(<Input type="password" />)}
          </Form.Item>
          <Form.Item
            name="confirm"
            label={t("Confirm")}
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("confirm", {
              rules: [
                {
                  required: true,
                },
              ],
            })(<Input type="password" />)}
          </Form.Item>
          <Form.Item
            wrapperCol={{ ...layoutButton.wrapperCol, offset: 8 }}
            labelCol={{ span: 6 }}
          >
            <Button type="primary" htmlType="submit">
              {t("Create New User")}
            </Button>
          </Form.Item>
        </Form>
      </Card>
    );
  }
}

export default compose(
  withTranslation(),
  Form.create({ name: "CreateUser" })
)(CreateUser);
