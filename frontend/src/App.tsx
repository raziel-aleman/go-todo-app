import "./App.css";
import useSWR from "swr";
// Import styles of packages that you've installed.
// All packages except `@mantine/hooks` require styles imports
import "@mantine/core/styles.css";
import { Box, List, MantineProvider, ThemeIcon } from "@mantine/core";
import AddTodo from "./components/AddTodo";
import { CheckCircleFillIcon } from "@primer/octicons-react";
import EditTodo from "./components/EditTodo";
import { redirect } from "react-router-dom";

export interface Todo {
	id: number;
	title: string;
	body: string;
	done: boolean;
}

export const ENDPOINT = "http://localhost:8080";

// const fetcher = (url: string) =>
// 	fetch(`${ENDPOINT}/${url}`, {
// 		method: "GET",
// 		credentials: 'include',
// 		// headers: {
// 		// 	"Cookie": document.cookie.split('; ').filter(row => row.startsWith('session_id=')).map(c=>c.split('=')[1])[0],
// 		// }
// 	}).then((res) => res.json());

const fetcher = async (url: string) => {
	const res = await fetch(`${ENDPOINT}/${url}`, {
		method: "GET",
		credentials: "include",
	});

	// If the status code is not in the range 200-299,
	// we redirect to login page to get new session
	if (!res.ok) {
		const error = new Error("An error occurred while fetching the data.");
		throw error;
	}

	return res.json();
};

export function App() {
	const { data, error, mutate } = useSWR<Todo[]>("api/todos", fetcher);

	if (error) {
		window.location.href = "http://localhost:3000/login";
	}

	const handleLogout = () => {
		window.location.href = "http://localhost:8080/auth/logout/github";
	};

	const markTodoAddDone = async (id: number) => {
		const updated = await fetch(`${ENDPOINT}/api/todos/${id}/done`, {
			method: "PATCH",
			credentials: "include",
			// headers: {
			// 	"Cookie": document.cookie.split('; ').filter(row => row.startsWith('session_id=')).map(c=>c.split('=')[1])[0],
			// }
		})
			.then((res) => res.json())
			.catch((err) => {
				console.log(err);
				redirect("localhost:3000/login");
			});

		mutate(updated);
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
				<List
					spacing="xs"
					size="sm"
					mb={12}
					center
					style={{ display: "inline-list-item" }}
				>
					{data?.map((todo) => {
						return (
							<div
								style={{
									display: "flex",
									alignItems: "center",
									justifyContent: "space-between",
									padding: "5px 0",
								}}
								key={`todo__list__${todo.id}`}
							>
								<List.Item
									style={{
										textAlign: "left",
										cursor: "pointer",
									}}
									onClick={() => markTodoAddDone(todo.id)}
									key={`todo__list__item__${todo.id}`}
									icon={
										todo.done ? (
											<ThemeIcon
												color="teal"
												size={24}
												radius="xl"
											>
												<CheckCircleFillIcon
													size={20}
												/>
											</ThemeIcon>
										) : (
											<ThemeIcon
												color="gray"
												size={24}
												radius="xl"
											>
												<CheckCircleFillIcon
													size={20}
												/>
											</ThemeIcon>
										)
									}
								>
									<span
										style={
											todo.done == true
												? {
														textDecorationLine:
															"line-through",
														color: "grey",
												}
												: {}
										}
									>
										{todo.title}
									</span>
									<br />
									<span
										style={{ color: "gray", fontSize: 12 }}
									>
										{todo.body}
									</span>
								</List.Item>
								<EditTodo mutate={mutate} data={todo} />
							</div>
						);
					})}
				</List>
				<AddTodo mutate={mutate} />
				<div style={{ paddingTop: "10px" }}>
					<button onClick={handleLogout}>Logout</button>
				</div>
			</Box>
		</MantineProvider>
	);
}

export default App;
