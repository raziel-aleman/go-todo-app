import { useState } from "react";
import { useForm } from "@mantine/form";
import {
  Button,
  Group,
  Modal,
  TextInput,
  Textarea,
  ThemeIcon,
} from "@mantine/core";
import { KeyedMutator } from "swr";
import { KebabHorizontalIcon } from "@primer/octicons-react";
import { ENDPOINT, Todo } from "../App";

interface Data {
  title: string;
  body: string;
  id: number;
  done: boolean;
}

interface EditTodoProps {
  mutate: KeyedMutator<Todo[]>;
  data: Data;
}

const EditTodo = ({ mutate, data }: EditTodoProps) => {
  const [open, setOpen] = useState(false);

  const form = useForm({
    initialValues: {
      title: data.title,
      body: data.body,
    },
  });

  const editTodo = async (values: { title: string; body: string }) => {
    console.log(JSON.stringify(values));
    const edited = await fetch(`${ENDPOINT}/api/todos/${data.id}/edit`, {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(values),
    }).then((res) => res.json());

    mutate(edited);
    setOpen(false);
  };

  return (
    <>
      <Modal opened={open} onClose={() => setOpen(false)} title="Edit todo">
        <form onSubmit={form.onSubmit(editTodo)}>
          <TextInput
            required
            mb={12}
            label="Todo"
            placeholder="What do you want to do?"
            {...form.getInputProps("title")}
          />
          <Textarea
            required
            mb={12}
            label="Body"
            placeholder="tell me more..."
            {...form.getInputProps("body")}
          />
          <Button type="submit"> Save changes </Button>
        </form>
      </Modal>

      <Group>
        <ThemeIcon
          color=""
          size={24}
          radius="xl"
          style={
            data.done
              ? { pointerEvents: "none", opacity: "0.4" }
              : { cursor: "pointer" }
          }
          onClick={() => {
            setOpen(true);
          }}
        >
          <KebabHorizontalIcon size={16} />
        </ThemeIcon>
      </Group>
    </>
  );
};

export default EditTodo;
