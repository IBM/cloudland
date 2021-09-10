/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Form, Card, Input, Button, message } from "antd";
import { createKeyApi } from "../../service/keys";
import "./keys.css";
const layoutButton = {
  labelCol: { span: 8 },
  wrapperCol: { span: 16 },
};
const layoutForm = {
  labelCol: { span: 6 },
  wrapperCol: { span: 10 },
  LayoutType: "horizontal",
};
class CreateKey extends Component {
  constructor(props) {
    super(props);
    this.state = {
      currentData: [],
    };
  }

  listKeys = () => {
    this.props.history.push("/keys");
  };
  handleSubmit = (e) => {
    console.log("handleSubmit:", e);
    e.preventDefault();
    this.props.form.validateFieldsAndScroll((err, values) => {
      if (!err) {
        console.log("handleSubmit-value:", values);

        createKeyApi(values)
          .then((res) => {
            this.props.history.push("/keys");
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
        title={"Create Key"}
        extra={
          <Button
            style={{ float: "right" }}
            type="primary"
            onClick={this.listKeys}
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
            label="Key Name"
            name="name"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("name", {
              rules: [
                {
                  required: true,
                },
              ],
            })(<Input />)}
          </Form.Item>
          <Form.Item
            name="pubkey"
            label="Public Key"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("pubkey", {})(
              <Input.TextArea
                showCount="true"
                autoSize={{ minRows: 3, maxRows: 6 }}
              />

            )}
          </Form.Item>
          <Form.Item
            wrapperCol={{ ...layoutButton.wrapperCol, offset: 8 }}
            labelCol={{ span: 6 }}
          >
            <Button type="primary" htmlType="submit">
              Create Key
            </Button>
          </Form.Item>
        </Form>
      </Card>
    );
  }
}
export default Form.create({ name: "CreateKey" })(CreateKey);
