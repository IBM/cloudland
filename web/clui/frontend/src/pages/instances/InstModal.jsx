/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Input, Select, Form, Modal } from "antd";
import { hypersListApi } from "../../service/hypers";
import { flavorsListApi } from "../../service/flavors";
const layoutForm = {
  labelCol: { span: 8 },
  wrapperCol: { span: 10 },
  layouttype: "horizontal",
};
const { Option } = Select;
// const loginInfor = JSON.parse(sessionStorage.loginInfo);

class InstModal extends Component {
  constructor(props) {
    super(props);
    this.state = {
      flavors: [],
      hypers: [],
      status: [],
      isLoaded: false,
    };
  }
  componentDidMount() {
    const _this = this;
    flavorsListApi()
      .then((res) => {
        _this.setState({
          flavors: res.flavors,
          isLoaded: true,
        });
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
    if (JSON.parse(sessionStorage.loginInfo).isAdmin) {
      hypersListApi()
        .then((res) => {
          _this.setState({
            hypers: res.hypers,
            isLoaded: true,
          });
        })
        .catch((error) => {
          _this.setState({
            isLoaded: false,
            error: error,
          });
        });
    }
  }
  handleOk = () => {
    const p = this;
    const { form } = this.props;
    form.validateFieldsAndScroll((err, values) => {
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
  getOptionFlavorList(data) {
    if (!data) {
      return [];
    }
    const options = [];
    data.map((item) => {
      options.push(
        <Option value={item.ID} key={item.ID}>
          {item.Name}
        </Option>
      );
    });
    return options;
  }
  getOptionHyperList(data) {
    if (!data) {
      return [];
    }
    const options = [];
    data.map((item, index) => {
      options.push(
        <Option value={index} key={item.ID}>
          {item.Hostname}
        </Option>
      );
    });
    return options;
  }

  initFormList = () => {
    const p = this;
    const { getFieldDecorator } = p.props.form;
    const { modalFormList } = this.props;
    const formItemList = [];
    if (modalFormList && modalFormList.length > 0) {
      modalFormList.forEach((item) => {
        const { label } = item;
        const { rules } = item;
        const rulesType = rules || [
          { required: true, message: `${label} required` },
        ];
        const initialValue = item.initialValue || undefined;

        const { placeholder } = item;
        const { width } = item;
        const { style } = item;
        const { name } = item;
        const { disabled } = item;
        const formItemLayout = {
          labelCol: {
            xs: { span: 24 },
            sm: { span: 6 },
          },
          wrapperCol: {
            xs: { span: 24 },
            sm: { span: 16 },
          },
        };

        if (item.field === this.props.title) {
          if (item.type === "INPUT") {
            const INPUT = (
              <Form.Item
                label={label}
                name={name}
                style={style}
                {...formItemLayout}
              >
                {getFieldDecorator(`${name}`, {
                  rules: rulesType,
                  initialValue,
                })(
                  <Input
                    type="text"
                    disabled={disabled || false}
                    style={{ width }}
                    // placeholder={placeholder}
                  />
                )}
              </Form.Item>
            );
            formItemList.push(INPUT);
          } else if (item.type === "SELECT") {
            if (name) {
              switch (name) {
                case "flavor": {
                  const SELECT = (
                    <Form.Item label={label} name={name} {...formItemLayout}>
                      {getFieldDecorator(`${name}`, {
                        initialValue, // true | false
                      })(
                        <Select style={{ width }} placeholder={placeholder}>
                          {this.getOptionFlavorList(this.state.flavors)}
                        </Select>
                      )}
                    </Form.Item>
                  );
                  return formItemList.push(SELECT);
                }
                case "hyper": {
                  const SELECT = (
                    <Form.Item
                      label={label}
                      name={name}
                      {...formItemLayout}
                      hidden={!JSON.parse(sessionStorage.loginInfo).isAdmin}
                    >
                      {getFieldDecorator(`${name}`, {
                        initialValue, // true | false
                      })(
                        <Select style={{ width }} placeholder={placeholder}>
                          {this.getOptionHyperList(this.state.hypers)}
                        </Select>
                      )}
                    </Form.Item>
                  );
                  return formItemList.push(SELECT);
                }
                case "action": {
                  const SELECT = (
                    <Form.Item label={label} name={name} {...formItemLayout}>
                      {getFieldDecorator(`${name}`, {
                        initialValue, // true | false
                      })(
                        <Select style={{ width }} placeholder={placeholder}>
                          <Option value="stop" key="stop">
                            stop
                          </Option>
                          <Option value="start" key="start">
                            start
                          </Option>
                        </Select>
                      )}
                    </Form.Item>
                  );
                  return formItemList.push(SELECT);
                }

                default:
                  return null;
              }
            }
          }
        }
      });
    }
    return formItemList;
  };

  render() {
    const p = this;

    return (
      <div>
        <Modal
          destroyOnClose
          title={this.props.title}
          visible={this.props.visible}
          onOk={this.handleOk}
          onCancel={this.handleCancel}
          maskClosable={false}
          okText={this.props.okText}
          cancelText={this.props.cancelText}
        >
          <Form wrapperCol={{ ...layoutForm.wrapperCol }}>
            {p.initFormList()}
          </Form>
        </Modal>
      </div>
    );
  }
}

export default Form.create({ name: "instModal" })(InstModal);
