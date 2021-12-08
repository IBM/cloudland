import React, { Component } from "react";
import { withTranslation } from "react-i18next";
import { compose } from "redux";
import { Form, Card, Input, Select, Button, message } from "antd";
import {
  createSecgroupApi,
  getSecgroupInforById,
  editSecgroupInfor,
} from "../../service/secgroups";
const layoutButton = {
  labelCol: { span: 8 },
  wrapperCol: { span: 16 },
};
const layoutForm = {
  labelCol: { span: 6 },
  wrapperCol: { span: 10 },
};
class ModifySecgroups extends Component {
  constructor(props) {
    super(props);
    this.state = {
      isShowEdit: false,
      currentData: [],
      isdefault: "no",
    };
    if (props.match.params.id) {
      getSecgroupInforById(props.match.params.id).then((res) => {
        this.setState({
          currentData: res,
          isShowEdit: true,
        });
      });
    }
  }
  listSecgroups = () => {
    this.props.history.push("/secgroups");
  };
  handleSubmit = (e) => {
    e.preventDefault();
    this.props.form.validateFieldsAndScroll((err, values) => {
      if (!err) {
        let params = {
          name: values.name,
          isdefault: values.isdefault + "",
        };
        if (this.props.match.params.id) {
          editSecgroupInfor(this.props.match.params.id, params).then((res) => {
            this.props.history.push("/secgroups");
          });
        } else {
          let params = {
            name: values.name,
            isdefault: values.isdefault,
          };
          createSecgroupApi(params)
            .then((res) => {
              this.props.history.push("/secgroups");
            })
            .catch((err) => {
              console.log("Error, create secgroup handleSubmit-error:", err);
            });
        }
      } else {
        message.error(" input wrong information");
      }
    });
  };
  render() {
    const { t } = this.props;
    return (
      <Card
        title={
          this.state.isShowEdit
            ? t("Edit Security Group")
            : t("Create New Security Group")
        }
        extra={
          <Button
            style={{
              float: "right",
              paddingLeft: "10px",
              paddingRight: "10px",
            }}
            type="primary"
            onClick={this.listSecgroups}
          >
            {t("Return")}
          </Button>
        }
      >
        <Form
          layout="horizontal"
          wrapperCol={{ ...layoutForm.wrapperCol }}
          onSubmit={(e) => this.handleSubmit(e)}
        >
          <Form.Item
            label={t("Name")}
            name="name"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("name", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.Name,
            })(<Input />)}
          </Form.Item>
          <Form.Item
            label={t("Is Default")}
            name="isdefault"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("isdefault", {
              rules: [],
              initialValue:
                this.state.currentData.IsDefault === true ? "yes" : "no",
            })(
              <Select placeholder={t("no")}>
                <Select.Option key="yes" value="yes">
                  {t("yes")}
                </Select.Option>
                <Select.Option key="no" value="no">
                  {t("no")}
                </Select.Option>
              </Select>
            )}
          </Form.Item>
          <Form.Item
            wrapperCol={{ ...layoutButton.wrapperCol, offset: 8 }}
            labelCol={{ span: 6 }}
          >
            {this.state.isShowEdit ? (
              <Button type="primary" htmlType="submit">
                {t("Update Security Group")}
              </Button>
            ) : (
              <Button type="primary" htmlType="submit">
                {t("Create New Security Group")}
              </Button>
            )}
          </Form.Item>
        </Form>
      </Card>
    );
  }
}

export default compose(
  withTranslation(),
  Form.create({ name: "modifySecgroups" })
)(ModifySecgroups);
