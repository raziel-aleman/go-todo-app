import { Navigate, Outlet } from "react-router-dom";

// const userSession = await fetch(`http://localhost:8080/auth/verify-session`, {
// 	method: "GET",
// 	credentials: 'include',
// 	headers: {
// 		"Cookie": document.cookie.split('; ').filter(row => row.startsWith('session_id=')).map(c=>c.split('=')[1])[0],
// 	}
// }).then((res) => res.json());

const ProtectedRoutes = () => {
	// TODO: Use authentication token
	//const userSession = document.cookie.split('; ').filter(row => row.startsWith('session_id=')).map(c=>c.split('=')[1])[0]

	//console.log(userSession)

	//if (userSession === undefined) {console.log("user session not found")} else {console.log("user session found")}

	return true ? <Outlet /> : <Navigate to="/login" replace />;
};

export default ProtectedRoutes;
