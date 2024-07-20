import { Navigate, Outlet } from "react-router-dom";

const ProtectedRoutes = () => {
	// TODO: Use authentication token
	const userSession = document.cookie.split('; ').filter(row => row.startsWith('user_session=')).map(c=>c.split('=')[1])[0]

	//if (userSession === undefined) {console.log("no user session")} else {console.log("user session found")}

	return userSession ? <Outlet /> : <Navigate to="/login" replace />;
};

export default ProtectedRoutes;
