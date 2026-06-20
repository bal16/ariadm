import { createSignal } from "solid-js";
// import "./App.css";

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
        <button
          type="button"
          class="border border-white-1 px-3 py-2 cursor-pointer"
          onClick={() => setCount((count) => count + 1)}
        >
          Count is {count()}
        </button>
      </section>
    </div>
  );
}

export default App;
