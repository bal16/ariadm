import { createSignal } from "solid-js";
import "./app.css";
import { Button } from "~/components/ui/button";

function App() {
  const [count, setCount] = createSignal(0);

  return (
    <div>
      <section class="">
        <div>
          <h1 class="text-4xl font-bold ">Hello World</h1>
          <p class="text-sm">
            Edit <code>src/App.tsx</code> and save to test <code>HMR</code>
          </p>
        </div>
        <Button type="button" onClick={() => setCount((count) => count + 1)}>
          Count is {count()}
        </Button>
      </section>
    </div>
  );
}

export default App;
