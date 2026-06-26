import { PORT } from "./config";

function App() {
  return (
    <>
      <section class="bg-gray-100 p-4 rounded-lg shadow-md dark:bg-gray-800 dark:text-white">
        <h3 class="text-xl font-bold">Ariadm Integration</h3>
        <p class="font-sm">Extension working and listen port {PORT}</p>
      </section>
    </>
  );
}

export default App;
