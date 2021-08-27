import React, { Component } from "react";
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
      isDefault: "",
    };
    if (props.match.params.id) {
      getSecgroupInforById(props.match.params.id).then((res) => {
        console.log("getSecgroupInforById:", res);
        this.setState({
          currentData: res,
          isShowEdit: true,
        });
        console.log("getSecgroupInforById-this.state:", this.state);
      });
    }
  }
  listSecgroups = () => {
    this.props.history.push("/secgroups");
  };
  handleSubmit = (e) => {
    console.log("handleSubmit:", e);
    e.preventDefault();
    this.props.form.validateFieldsAndScroll((err, values) => {
      if (!err) {
        console.log("handleSubmit-value:", values);
        console.log("提交");
        if (this.props.match.params.id) {
          //const _this = this;
          editSecgroupInfor(this.props.match.params.id, values).then((res) => {
            console.log("editSecgroupInfor:", res);
            // _this.setState({
            //   isShowEdit: ! this.state.isShowEdit,
            // });
            this.props.history.push("/secgroups");
          });
        } else {
          console.log("before-createSecgroupApi:", values);
          values.isdefault =
            values.isdefault === undefined
              ? this.state.isDefault
              : values.isdefault;
          createSecgroupApi(values)
            .then((res) => {
              console.log("handleSubmit-res-createSecgroupApi:", res);
              this.props.history.push("/secgroups");
            })
            .catch((err) => {
              console.log("handleSubmit-error:", err);
            });
        }
      } else {
        message.error(" input wrong information");
      }
    });
  };
  render() {
    return (
      <Card
        title={
          this.state.isShowEdit
            ? "Edit Security Group"
            : "Create New Security Group "
        }
        extra={
          <Button
            style={{
              float: "right",
              "padding-left": "10px",
              "padding-right": "10px",
            }}
            type="primary"
            onClick={this.listSecgroups}
          >
            Return
          </Button>
        }
      >
        <Form
          layout="horizontal"
          wrapperCol={{ ...layoutForm.wrapperCol }}
          onSubmit={(e) => this.handleSubmit(e)}
        >
          <Form.Item
            label="Name"
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
            label="Is Default"
            name="isdefault"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("isdefault", {
              rules: [],
              initialValue: this.state.currentData.isDefault,
            })(
              <Select placeholder="no">
                <Select.Option key="yes" value="yes">
                  yes
                </Select.Option>
                <Select.Option key="no" value="no">
                  no
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
                Update Security Group
              </Button>
            ) : (
              <Button type="primary" htmlType="submit">
                Create Security Group
              </Button>
            )}
          </Form.Item>
        </Form>
      </Card>
    );
  }
}
export default Form.create({ name: "modifySecgroups" })(ModifySecgroups);
