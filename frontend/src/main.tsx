import React from "react";
import ReactDOM from "react-dom/client";
import {
	createBrowserRouter,
	RouterProvider,
} from "react-router-dom";
import App from "./App.tsx";
import "./index.css";
import Login from "./components/Login.tsx";
import ProtectedRoutes from "./components/ProtectedRoutes.tsx";

const router = createBrowserRouter(
	[
		// {
		// 	path: "/",
		// 	element: <Login />,
		// },
		{
			path: "/login",
			element: <Login />,
		},
		{
			element: <ProtectedRoutes />,
			children: [
				{
					path: "/*",
					element: <App />,
				},
				// {
				// 	path: "/route2",
				// 	element: <Screen2 />,
				// },
				// {
				// 	path: "/route3",
				// 	element: <Screen3 />,
				// },
			],
		},
	],
	// { basename: "/app" }
);

ReactDOM.createRoot(document.getElementById("root")!).render(
	<React.StrictMode>
		<RouterProvider router={router} />
	</React.StrictMode>
);
