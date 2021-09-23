import React, { Component } from "react";
import { Table } from "antd";
import { withTranslation } from "react-i18next";
class DataTable extends Component {
  render() {
    const { t } = this.props;
    return (
      <Table
        rowKey={this.props.rowKey}
        columns={this.props.columns}
        dataSource={this.props.dataSource}
        bordered={this.props.bordered}
        pagination={{
          total: this.props.total, //total count
          defaultPageSize: this.props.pageSize, //default pageSize
          showSizeChanger: true,

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
