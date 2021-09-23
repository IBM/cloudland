/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Form, Card, Input, Button, message } from "antd";
import { createFlavorApi } from "../../service/flavors";
import { withTranslation } from "react-i18next";
import { compose } from "redux";
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
    this.state = {
      name: "",
      cpu: 1,
      memory: 1,
      disk: 1,
      swap: 0,
      ephemeral: 0,
    };
  }

  listFlavors = () => {
    this.props.history.push("/flavors");
  };
  handleSubmit = (event) => {
    event.preventDefault();
    this.props.form.validateFieldsAndScroll((err) => {
      if (!err) {
        createFlavorApi({
          Name: this.state.name,
          Cpu: this.state.cpu,
          Memory: parseInt(this.state.memory),
          Disk: parseInt(this.state.disk),
          Swap: parseInt(this.state.swap),
          Ephemeral: parseInt(this.state.ephemeral),
        })
          .then((res) => {
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
    const { t } = this.props;
    return (
      <Card
        title={t("Create New Flavor")}
        extra={
          <Button
            style={{ float: "right" }}
            type="primary"
            onClick={this.listFlavors}
          >
            {t("Return")}
          </Button>
        }
      >
        <Form
          layout="horizontal"
          onSubmit={(e) => this.handleSubmit(e)}
          wrapperCol={{ ...layoutForm.wrapperCol }}
        >
          <Form.Item
            label={t("Name")}
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
            label={t("Cpu")}
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
            label={t("Memory") + "(M)"}
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
            label={t("Disk") + "(G)"}
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
            label={t("Swap") + "(G)"}
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
            label={t("Ephemeral") + "(G)"}
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
                {t("Create New Flavor")}
              </Button>
            }
          </Form.Item>
        </Form>
      </Card>
    );
  }
}

export default compose(
  withTranslation(),
  Form.create({ name: "createFlavors" })
)(CreateFlavors);
