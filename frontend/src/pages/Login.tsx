const Login = () => {
  const handleLogin = () => {
    window.location.href = "http://localhost:8080/auth/github";
  };

  return (
    <div>
      <button onClick={handleLogin}>Login with Github</button>
    </div>
  );
};

export default Login;
