/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import Frame from "./components/Frame/Frame";
import { Switch, Route, Redirect } from "react-router-dom";
import { connect } from "react-redux";
import "antd/dist/antd.css";
import "./App.css";
import { mainRoutes } from "./routes";
import { isLogined } from "./utils/auth";
function App() {
  return isLogined() ? (
    <Frame>
      <Switch>
        {mainRoutes.map((route) => {
          return (
            <Route
              key={route.path}
              path={route.path}
              exact={route.exact}
              render={(routeProps) => {
                return <route.component {...routeProps} />;
              }}
            />
          );
        })}
        <Redirect to={mainRoutes[0].path} from="/" />
        <Redirect to="/404" />
      </Switch>
    </Frame>
  ) : (
    <Redirect to="/login" />
  );
}

export default App;
// export default App;
