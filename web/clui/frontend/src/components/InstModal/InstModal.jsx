/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Input, Select, Form, Modal } from "antd";
import { hypersListApi } from "../../api/hypers";
import Utils from "../../pages/instances/Instances";
import { flavorsListApi } from "../../api/flavors";
const layoutForm = {
  labelCol: { span: 8 },
  wrapperCol: { span: 10 },
  LayoutType: "horizontal",
};
const { Option } = Select;
class InstModal extends Component {
  constructor(props) {
    super(props);
    this.state = {
      flavors: [],
      hypers: [],
      isLoaded: false,
    };
  }
  componentDidMount() {
    this.props.onRef(this);
    const _this = this;
    flavorsListApi()
      .then((res) => {
        _this.setState({
          flavors: res.flavors,
          isLoaded: true,
        });
        console.log("flavors:", res);
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
    hypersListApi()
      .then((res) => {
        _this.setState({
          hypers: res.hypers,
          isLoaded: true,
        });
        console.log("hypersListApi~~", this.state.hypers);
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
  }
  handleOk = () => {
    const p = this;
    const { form } = this.props;
    console.log("handleOk-form", this.props);
    form.validateFieldsAndScroll((err, values) => {
      console.log("handleOk", values);
      if (err) {
        return;
      }
      // this.state.fileList?values.image = this.state.fileList : []
      p.props.submit(values);
    });
  };

  handleCancel = () => {
    const { close } = this.props;

    close();
  };

  getOptionList(data) {
    console.log("getOptionList", data);
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
    console.log("getOptionList", data);
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
    console.log("basic-modalFormList", modalFormList);
    const formItemList = [];
    if (modalFormList && modalFormList.length > 0) {
      modalFormList.forEach((item) => {
        console.log("modalFormList-item", item);
        const { label } = item;
        const { rules } = item;
        const rulesType = rules || [
          { required: true, message: `${label}必填` },
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
          } else if (item.type === "SELECT" && name === "flavor") {
            const SELECT = (
              <Form.Item label={label} name={name} {...formItemLayout}>
                {getFieldDecorator(`${name}`, {
                  initialValue, // true | false
                })(
                  <Select style={{ width }} placeholder={placeholder}>
                    {this.getOptionList(this.state.flavors)}
                  </Select>
                )}
              </Form.Item>
            );
            formItemList.push(SELECT);
          } else {
            const SELECT = (
              <Form.Item label={label} name={name} {...formItemLayout}>
                {getFieldDecorator(`${name}`, {
                  initialValue, // true | false
                })(
                  <Select style={{ width }} placeholder={placeholder}>
                    {this.getOptionHyperList(this.state.hypers)}
                  </Select>
                )}
              </Form.Item>
            );
            formItemList.push(SELECT);
          }
        }
      });
      return formItemList;
    }
  };

  render() {
    const p = this;
    console.log("instModal-key", this.props.key);
    // const { getFieldDecorator } = this.props.form;
    return (
      <div>
        <Modal
          destroyOnClose
          title={this.props.title}
          visible={this.props.visible}
          onOk={this.handleOk}
          onCancel={this.handleCancel}
          maskClosable={false}
        >
          <Form
            onSubmit={(e) => {
              console.log("modal-e", e);
              // this.handleModalSubmit(e);
            }}
            //   layout={{ ...layoutForm.LayoutType }}
            wrapperCol={{ ...layoutForm.wrapperCol }}
          >
            {p.initFormList()}
          </Form>
        </Modal>
      </div>
    );
  }
}

export default Form.create({ name: "instModal" })(InstModal);