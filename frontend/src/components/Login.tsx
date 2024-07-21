import { Box, Button, MantineProvider, Text } from "@mantine/core";

const Login = () => {
	const handleLogin = () => {
		window.location.href = "http://localhost:8080/auth/github";
	};

	return (
		<MantineProvider>
			<Box
				bg="gray.8"
				style={{
					padding: "2rem",
					width: "100%",
					maxWidth: "40rem",
					margin: "0 auto",
					borderRadius: "20px",
				}}
			>
				<Text c="white" p="sm" size="xl">
					Welcome to GoDo!
				</Text>
				{/* <button onClick={handleLogin}>Login with Github</button> */}
				<Button onClick={handleLogin} m="lg">
					<Text size="lg">Login with Github</Text>
				</Button>
			</Box>
		</MantineProvider>
	);
};

export default Login;
