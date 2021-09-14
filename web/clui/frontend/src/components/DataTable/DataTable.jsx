import React, { Component } from "react";
import { Table } from "antd";
import { withTranslation } from "react-i18next";
class DataTable extends Component {
  constructor(props) {
    super(props);
    console.log("this.props--dataTable", this.props);
  }
  render() {
    const { t } = this.props;
    return (
      <Table
        rowKey={this.props.rowKey}
        columns={this.props.columns}
        dataSource={this.props.dataSource}
        bordered={this.props.bordered}
        pagination={{
          //pagination
          total: this.props.total, //total count
          defaultPageSize: this.props.pageSize, //default pageSize
          showSizeChanger: true, //是否显示可以设置几条一页的选项
          // onShowSizeChange: (current, pageSize) => {
          //   console.log("onShowSizeChange:", current, pageSize);
          //   //当几条一页的值改变后调用函数，current：改变显示条数时当前数据所在页；pageSize:改变后的一页显示条数
          //   this.props.toSelectchange(current, pageSize);
          // },
          onShowSizeChange: this.props.onShowSizeChange,
          onChange: this.props.onPaginationChange,
          showTotal: () => {
            return t("Total") + this.props.total + t("Items");
          },
          pageSizeOptions: this.props.pageSizeOptions,
        }}
        scroll={this.props.scroll}
        loading={this.props.loading}
        onHeaderRow={this.props.onHeaderRow}
      ></Table>
    );
  }
}
export default withTranslation()(DataTable);
