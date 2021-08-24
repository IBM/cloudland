/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Form, Card, Input, Select, Button, message } from "antd";
import {
  getUserInforById,
  editUserInfor,
} from "../../api/users";
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
class ModifyUser extends Component {
  constructor(props) {
    super(props);
    console.log("ModifyUser~~", props);
    this.state = {
      isShowEdit: false,
      currentData: [],
    };

    getUserInforById(props.match.params.id).then((res) => {
      console.log("getUserInforById:", res);
      this.setState({
        currentData: res,
        isShowEdit: true,
      });
    });
    
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
          //const _this = this;
        editUserInfor(this.props.match.params.id, values).then((res) => {
            console.log("editUserInfor:", res);
        
            this.props.history.push("/users");
          });
        
      } else {
        message.error(" input wrong information");
      }
    });
  };

  render() {
    return (
      <Card
        title={"Edit Registry"}
        extra={
          <Button type="primary" onClick={this.listUsers}>
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
            label="Password"
            name="password"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("password", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.Password,
            })(<Input />)}
          </Form.Item>
          <Form.Item
            label="Organizations"
            name="members"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("members", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.Members,
            })(
              <Select
                multiple={true}
              >
                {
	              this.state.currentData.Members.map((item,i)=>{
		            return(
			         <Select.Option key={i} value={item.OrgName}>{item.OrgName}</Select.Option>
		            )
	              }
	              )
                }
                
              </Select>
              
            )}
          </Form.Item>

          <Form.Item
            wrapperCol={{ ...layoutButton.wrapperCol, offset: 8 }}
            labelCol={{ span: 6 }}
          >
            {
	          <Button type="primary" htmlType="submit">
                Update Registry
              </Button>
            }
          </Form.Item>
        </Form>
      </Card>
    );
  }
}
export default Form.create({ name: "modifyUser" })(ModifyUser);
