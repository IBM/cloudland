<!-- page content -->
<div class="right_col" role="main">
    <div class="">
        <div class="page-title">
            <div class="title_left">
                <h3>Compute</h3>
            </div>
            <div class="title_right">
                <div class="col-md-5 col-sm-5 col-xs-12 form-group pull-right top_search">
                    <div class="input-group">
                        <input type="text" class="form-control" placeholder="Search for...">
                        <span class="input-group-btn">
                            <button class="btn btn-default" type="button">Go!</button>
                        </span>
                    </div>
                </div>
            </div>
        </div>
        <div class="clearfix"></div>
        <div class="row">
            <div class="col-md-12">
                <div class="x_panel">
                    <div class="x_title">
                        <h2>Compute Resource Usage</h2>
                        <ul class="nav navbar-right panel_toolbox">
                            <li><a class="collapse-link"><i class="fa fa-chevron-up"></i></a></li>
                        </ul>
                        <div class="clearfix"></div>
                    </div>
                    <div class="x_content">
                        <div class="row">
                            <div class="col-md-12">
                                <div class="panel panel-body">
                                    <div class="col-md-12">
                                        <div class="row">
                                            <div class="col-xs-3">
                                                <span class="chart" data-percent="73">
                                                    <span class="percent"></span>
                                                </span>
                                            </div>
                                            <div class="col-xs-3">
                                                <span class="chart" data-percent="86">
                                                    <span class="percent"></span>
                                                </span>
                                            </div>
                                            <div class="col-xs-3">
                                                <span class="chart" data-percent="73">
                                                    <span class="percent"></span>
                                                </span>
                                            </div>
                                            <div class="col-xs-3">
                                                <span class="chart" data-percent="60">
                                                    <span class="percent"></span>
                                                </span>
                                            </div>
                                            <div class="clearfix"></div>
                                        </div>
                                    </div>
                                    <div class="col-md-12">
                                        <div class="row">
                                            <div class="col-xs-3">
                                                <span>
                                                    <p style="padding:10px 0px 0px 15px;"> Instances
                                                        usage</p>
                                                </span>
                                            </div>
                                            <div class="col-xs-3">
                                                <span>
                                                    <p style="padding:10px 0px 0px 15px;"> CPU usage</p>
                                                </span>
                                            </div>
                                            <div class="col-xs-3">
                                                <span>
                                                    <p style="padding:10px 0px 0px 15px;">Memory usage
                                                    </p>
                                                </span>
                                            </div>
                                            <div class="col-xs-3">
                                                <span>
                                                    <p style="padding:10px 0px 0px 15px;">Storage usage
                                                    </p>
                                                </span>
                                            </div>
                                            <div class="clearfix"></div>
                                        </div>
                                    </div>
                                    <div class="col-md-12">
                                        <div class="row">
                                            <div class="col-xs-3">
                                                <span>
                                                    <p style="padding:1px 10px 0px 35px;">5/10</p>
                                                </span>
                                            </div>
                                            <div class="col-xs-3">
                                                <span>
                                                    <p style="padding:1px 0px 0px 35px;">6/10</p>
                                                </span>
                                            </div>
                                            <div class="col-xs-3">
                                                <span>
                                                    <p style="padding:1px 0px 0px 35px;">7/10</p>
                                                </span>
                                            </div>
                                            <div class="col-xs-3">
                                                <span>
                                                    <p style="padding:1px 0px 0px 35px;">8/10</p>
                                                </span>
                                            </div>
                                            <div class="clearfix"></div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        <!-- /list instances content -->
        <div class="row">
            <div class="col-md-12 col-sm-12 col-xs-12">
                <div class="x_panel">
                    <div class="x_title">
                        <h2 style="padding:0px 30px 0px 0px;">Instances</h2>
                        <button type="button" class="navbar-right btn btn-primary">Create</button>
                        <div class="clearfix"></div>
                    </div>
                    <div class="x_content">
                        <table id="datatable-checkbox"
                            class="table table-striped table-bordered bulk_action jambo_table">
                            <thead>
                                <tr>
                                    <th>
                                    <th><input type="checkbox" id="check-all" class="flat"></th>
                                    </th>
                                    <th>ID</th>
                                    <th>Hostname</th>
                                    <th>IP Address</th>
                                    <th>Image</th>
                                    <th>Status</th>
                                    <th>Owner</th>
                                    <th>Action</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr>
                                    <td>
                                    <th><input type="checkbox" id="check-all" class="flat"></th>
                                    </td>
                                    <td>934538536147244</td>
                                    <td>Instance-1</td>
                                    <td>192.168.1.1/24</td>
                                    <td>Ubuntu-mini</td>
                                    <td><span class="label label-success">Running</span></td>
                                    <td>Spark</td>
                                    <td>
                                      <div class="btn-group">
                                        <button data-toggle="dropdown" class="btn btn-primary dropdown-toggle btn-sm" type="button" aria-expanded="false">Action <span class="caret"></span>
                                        </button>
                                        <ul role="menu" class="dropdown-menu">
                                          <li><a href="#">View</a>
                                          </li>
                                          <li><a href="#">Stop</a>
                                          </li>
                                          <li><a href="#">Suspend</a>
                                          </li>
                                          <li class="divider"></li>
                                          <li><a href="#">Delete</a>
                                          </li>
                                        </ul>
                                      </div>
                                    </td>
                                </tr>
                                <tr>
                                    <td>
                                    <th><input type="checkbox" id="check-all" class="flat"></th>
                                    </td>
                                    <td>934538536147244</td>
                                    <td>Instance-2</td>
                                    <td>192.168.1.2/24</td>
                                    <td>Ubuntu-mini</td>
                                    <td><span class="label label-danger">Error</span></td>
                                    <td>Spark</td>
                                    <td>
                                      <div class="btn-group">
                                        <button data-toggle="dropdown" class="btn btn-primary dropdown-toggle btn-sm" type="button" aria-expanded="false">Action <span class="caret"></span>
                                        </button>
                                        <ul role="menu" class="dropdown-menu">
                                          <li><a href="#">View</a>
                                          </li>
                                          <li><a href="#">Stop</a>
                                          </li>
                                          <li><a href="#">Suspend</a>
                                          </li>
                                          <li class="divider"></li>
                                          <li><a href="#">Delete</a>
                                          </li>
                                        </ul>
                                      </div>
                                    </td>
                                </tr>
                                <tr>
                                    <td>
                                    <th><input type="checkbox" id="check-all" class="flat"></th>
                                    </td>
                                    <td>934538536147244</td>
                                    <td>Instance-3</td>
                                    <td>192.168.1.3/24</td>
                                    <td>Ubuntu-mini</td>
                                    <td><span class="label label-warning">Stopped</span></td>
                                    <td>Spark</td>
                                    <td>
                                      <div class="btn-group">
                                        <button data-toggle="dropdown" class="btn btn-primary dropdown-toggle btn-sm" type="button" aria-expanded="false">Action <span class="caret"></span>
                                        </button>
                                        <ul role="menu" class="dropdown-menu">
                                          <li><a href="#">View</a>
                                          </li>
                                          <li><a href="#">Stop</a>
                                          </li>
                                          <li><a href="#">Suspend</a>
                                          </li>
                                          <li class="divider"></li>
                                          <li><a href="#">Delete</a>
                                          </li>
                                        </ul>
                                      </div>
                                    </td>
                                </tr>
                                <tr>
                                    <td>
                                    <th><input type="checkbox" id="check-all" class="flat"></th>
                                    </td>
                                    <td>934538536147244</td>
                                    <td>Instance-4</td>
                                    <td>192.168.1.4/24</td>
                                    <td>Ubuntu-mini</td>
                                    <td><span class="label label-default">Deleteing</span></td>
                                    <td>Spark</td>
                                    <td>
                                      <div class="btn-group">
                                        <button data-toggle="dropdown" class="btn btn-primary dropdown-toggle btn-sm" type="button" aria-expanded="false">Action <span class="caret"></span>
                                        </button>
                                        <ul role="menu" class="dropdown-menu">
                                          <li><a href="#">View</a>
                                          </li>
                                          <li><a href="#">Stop</a>
                                          </li>
                                          <li><a href="#">Suspend</a>
                                          </li>
                                          <li class="divider"></li>
                                          <li><a href="#">Delete</a>
                                          </li>
                                        </ul>
                                      </div>
                                    </td>
                                </tr>
                                <tr>
                                    <td>
                                    <th><input type="checkbox" id="check-all" class="flat"></th>
                                    </td>
                                    <td>934538536147244</td>
                                    <td>Instance-4</td>
                                    <td>192.168.1.4/24</td>
                                    <td>Ubuntu-mini</td>
                                    <td><span class="label label-success">Running</span></td>
                                    <td>Spark</td>
                                    <td>
                                      <div class="btn-group">
                                        <button data-toggle="dropdown" class="btn btn-primary dropdown-toggle btn-sm" type="button" aria-expanded="false">Action <span class="caret"></span>
                                        </button>
                                        <ul role="menu" class="dropdown-menu">
                                          <li><a href="#">View</a>
                                          </li>
                                          <li><a href="#">Stop</a>
                                          </li>
                                          <li><a href="#">Suspend</a>
                                          </li>
                                          <li class="divider"></li>
                                          <li><a href="#">Delete</a>
                                          </li>
                                        </ul>
                                      </div>
                                    </td>
                                </tr>
                                <tr>
                                    <td>
                                    <th><input type="checkbox" id="check-all" class="flat"></th>
                                    </td>
                                    <td>934538536147244</td>
                                    <td>Instance-4</td>
                                    <td>192.168.1.4/24</td>
                                    <td>Ubuntu-mini</td>
                                    <td><span class="label label-success">Running</span></td>
                                    <td>Spark</td>
                                    <td>
                                      <div class="btn-group">
                                        <button data-toggle="dropdown" class="btn btn-primary dropdown-toggle btn-sm" type="button" aria-expanded="false">Action <span class="caret"></span>
                                        </button>
                                        <ul role="menu" class="dropdown-menu">
                                          <li><a href="#">View</a>
                                          </li>
                                          <li><a href="#">Stop</a>
                                          </li>
                                          <li><a href="#">Suspend</a>
                                          </li>
                                          <li class="divider"></li>
                                          <li><a href="#">Delete</a>
                                          </li>
                                        </ul>
                                      </div>
                                    </td>
                                </tr>
                                <tr>
                                    <td>
                                    <th><input type="checkbox" id="check-all" class="flat"></th>
                                    </td>
                                    <td>934538536147244</td>
                                    <td>Instance-4</td>
                                    <td>192.168.1.4/24</td>
                                    <td>Ubuntu-mini</td>
                                    <td> <span class="label label-info">Creating</span></td>
                                    <td>Spark</td>
                                    <td>
                                      <div class="btn-group">
                                        <button data-toggle="dropdown" class="btn btn-primary dropdown-toggle btn-sm" type="button" aria-expanded="false">Action <span class="caret"></span>
                                        </button>
                                        <ul role="menu" class="dropdown-menu">
                                          <li><a href="#">View</a>
                                          </li>
                                          <li><a href="#">Stop</a>
                                          </li>
                                          <li><a href="#">Suspend</a>
                                          </li>
                                          <li class="divider"></li>
                                          <li><a href="#">Delete</a>
                                          </li>
                                        </ul>
                                      </div>
                                    </td>
                                </tr>
                                <tr>
                                    <td>
                                    <th><input type="checkbox" id="check-all" class="flat"></th>
                                    </td>
                                    <td>934538536147244</td>
                                    <td>Instance-4</td>
                                    <td>192.168.1.4/24</td>
                                    <td>Ubuntu-mini</td>
                                    <td><span class="label label-success">Running</span></td>
                                    <td>Spark</td>
                                    <td>
                                      <div class="btn-group">
                                        <button data-toggle="dropdown" class="btn btn-primary dropdown-toggle btn-sm" type="button" aria-expanded="false">Action <span class="caret"></span>
                                        </button>
                                        <ul role="menu" class="dropdown-menu">
                                          <li><a href="#">View</a>
                                          </li>
                                          <li><a href="#">Stop</a>
                                          </li>
                                          <li><a href="#">Suspend</a>
                                          </li>
                                          <li class="divider"></li>
                                          <li><a href="#">Delete</a>
                                          </li>
                                        </ul>
                                      </div>
                                    </td>
                                </tr>
                                <tr>
                                    <td>
                                    <th><input type="checkbox" id="check-all" class="flat"></th>
                                    </td>
                                    <td>934538536147244</td>
                                    <td>Instance-4</td>
                                    <td>192.168.1.4/24</td>
                                    <td>Ubuntu-mini</td>
                                    <td><span class="label label-success">Running</span></td>
                                    <td>Spark</td>
                                    <td>
                                      <div class="btn-group">
                                        <button data-toggle="dropdown" class="btn btn-primary dropdown-toggle btn-sm" type="button" aria-expanded="false">Action <span class="caret"></span>
                                        </button>
                                        <ul role="menu" class="dropdown-menu">
                                          <li><a href="#">View</a>
                                          </li>
                                          <li><a href="#">Stop</a>
                                          </li>
                                          <li><a href="#">Suspend</a>
                                          </li>
                                          <li class="divider"></li>
                                          <li><a href="#">Delete</a>
                                          </li>
                                        </ul>
                                      </div>
                                    </td>
                                </tr>
                                <tr>
                                    <td>
                                    <th><input type="checkbox" id="check-all" class="flat"></th>
                                    </td>
                                    <td>934538536147244</td>
                                    <td>Instance-4</td>
                                    <td>192.168.1.4/24</td>
                                    <td>Ubuntu-mini</td>
                                    <td><span class="label label-success">Running</span></td>
                                    <td>Spark</td>
                                    <td>
                                      <div class="btn-group">
                                        <button data-toggle="dropdown" class="btn btn-primary dropdown-toggle btn-sm" type="button" aria-expanded="false">Action <span class="caret"></span>
                                        </button>
                                        <ul role="menu" class="dropdown-menu">
                                          <li><a href="#">View</a>
                                          </li>
                                          <li><a href="#">Stop</a>
                                          </li>
                                          <li><a href="#">Suspend</a>
                                          </li>
                                          <li class="divider"></li>
                                          <li><a href="#">Delete</a>
                                          </li>
                                        </ul>
                                      </div>
                                    </td>
                                </tr>
                                <tr>
                                    <td>
                                    <th><input type="checkbox" id="check-all" class="flat"></th>
                                    </td>
                                    <td>934538536147244</td>
                                    <td>Instance-4</td>
                                    <td>192.168.1.4/24</td>
                                    <td>Ubuntu-mini</td>
                                    <td><span class="label label-success">Running</span></td>
                                    <td>Spark</td>
                                    <td>
                                      <div class="btn-group">
                                        <button data-toggle="dropdown" class="btn btn-primary dropdown-toggle btn-sm" type="button" aria-expanded="false">Action <span class="caret"></span>
                                        </button>
                                        <ul role="menu" class="dropdown-menu">
                                          <li><a href="#">View</a>
                                          </li>
                                          <li><a href="#">Stop</a>
                                          </li>
                                          <li><a href="#">Suspend</a>
                                          </li>
                                          <li class="divider"></li>
                                          <li><a href="#">Delete</a>
                                          </li>
                                        </ul>
                                      </div>
                                    </td>
                                </tr>
                                <tr>
                                    <td>
                                    <th><input type="checkbox" id="check-all" class="flat"></th>
                                    </td>
                                    <td>934538536147244</td>
                                    <td>Instance-5</td>
                                    <td>192.168.1.5/24</td>
                                    <td>Ubuntu-mini</td>
                                    <td><span class="label label-success">Running</span></td>
                                    <td>Spark</td>
                                    <td>
                                      <div class="btn-group">
                                        <button data-toggle="dropdown" class="btn btn-primary dropdown-toggle btn-sm" type="button" aria-expanded="false">Action <span class="caret"></span>
                                        </button>
                                        <ul role="menu" class="dropdown-menu">
                                          <li><a href="#">View</a>
                                          </li>
                                          <li><a href="#">Stop</a>
                                          </li>
                                          <li><a href="#">Suspend</a>
                                          </li>
                                          <li class="divider"></li>
                                          <li><a href="#">Delete</a>
                                          </li>
                                        </ul>
                                      </div>
                                    </td>
                                </tr>
                                <tr>
                                    <td>
                                    <th><input type="checkbox" id="check-all" class="flat"></th>
                                    </td>
                                    <td>934538536147244</td>
                                    <td>Instance-5</td>
                                    <td>192.168.1.5/24</td>
                                    <td>Ubuntu-mini</td>
                                    <td><span class="label label-success">Running</span></td>
                                    <td>Spark</td>
                                    <td>
                                      <div class="btn-group">
                                        <button data-toggle="dropdown" class="btn btn-primary dropdown-toggle btn-sm" type="button" aria-expanded="false">Action <span class="caret"></span>
                                        </button>
                                        <ul role="menu" class="dropdown-menu">
                                          <li><a href="#">View</a>
                                          </li>
                                          <li><a href="#">Stop</a>
                                          </li>
                                          <li><a href="#">Suspend</a>
                                          </li>
                                          <li class="divider"></li>
                                          <li><a href="#">Delete</a>
                                          </li>
                                        </ul>
                                      </div>
                                    </td>
                                </tr>
                                <tr>
                                    <td>
                                    <th><input type="checkbox" id="check-all" class="flat"></th>
                                    </td>
                                    <td>934538536147244</td>
                                    <td>Instance-5</td>
                                    <td>192.168.1.5/24</td>
                                    <td>Ubuntu-mini</td>
                                    <td><span class="label label-success">Running</span></td>
                                    <td>Spark</td>
                                    <td>
                                      <div class="btn-group">
                                        <button data-toggle="dropdown" class="btn btn-primary dropdown-toggle btn-sm" type="button" aria-expanded="false">Action <span class="caret"></span>
                                        </button>
                                        <ul role="menu" class="dropdown-menu">
                                          <li><a href="#">View</a>
                                          </li>
                                          <li><a href="#">Stop</a>
                                          </li>
                                          <li><a href="#">Suspend</a>
                                          </li>
                                          <li class="divider"></li>
                                          <li><a href="#">Delete</a>
                                          </li>
                                        </ul>
                                      </div>
                                    </td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>
<!-- /page content -->
