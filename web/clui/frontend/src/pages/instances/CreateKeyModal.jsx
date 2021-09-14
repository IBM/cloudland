/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Input, Form, Modal } from "antd";
import { withTranslation } from "react-i18next";
import { compose } from "redux";
const modalFormItem = {
  labelCol: {
    xs: { span: 24 },
    sm: { span: 6 },
  },
  wrapperCol: {
    xs: { span: 24 },
    sm: { span: 16 },
  },
};
class CreateKeyModal extends Component {
  constructor(props) {
    super(props);
    this.state = {
      flavors: [],
      hypers: [],
      status: [],
      isLoaded: false,
    };
  }

  handleOk = () => {
    const p = this;
    const { form } = this.props;
    console.log("handleOk-key-form", this.props);
    form.validateFieldsAndScroll((err, values) => {
      console.log("handleOk-key", values);
      if (err) {
        return;
      }
      p.props.submit(values);
    });
  };

  handleCancel = () => {
    const { close } = this.props;
    close();
  };

  render() {
    const { t } = this.props;
    return (
      <div>
        <Modal
          destroyOnClose
          title={this.props.title}
          visible={this.props.visible}
          onOk={this.handleOk}
          onCancel={this.handleCancel}
          maskClosable={false}
          okText={t("OK")}
          cancelText={t("Cancel")}
        >
          <Form>
            <Form.Item label={t("Name")} name="name" {...modalFormItem}>
              {this.props.form.getFieldDecorator("name", {
                rules: [
                  {
                    required: true,
                  },
                ],
              })(<Input />)}
            </Form.Item>
            <Form.Item name="pubkey" label={t("Public Key")} {...modalFormItem}>
              {this.props.form.getFieldDecorator("pubkey", {
                rules: [
                  {
                    required: true,
                  },
                ],
              })(<Input.TextArea autoSize={{ minRows: 3, maxRows: 6 }} />)}
            </Form.Item>
          </Form>
        </Modal>
      </div>
    );
  }
}

export default compose(
  withTranslation(),
  Form.create({ name: "createKeyModal" })
)(CreateKeyModal);
