'use client';

import { useState, useEffect } from "react";

export default function Home() {
  const [greeting, setGreeting] = useState('');

  // Using Wails bindings must be done in a `useEffect` because they need the global `window` object
  useEffect(() => {
    (async () =>  {
      // Wails bindings imports must be dynamic also because importing bindings relies on the
      // `window` object too
      const Greet = (await import("@/bindings/main/GreetService")).Greet;
      // Output values from the Wails bindings to be displayed should be placed into in some kind of
      // state variable
      setGreeting(await Greet('wallets list'));
    })();
  }, []);

  return (
    <main className="prose bg-base h-[100vh] max-w-none grid place-content-center">
      <h1 className="text-5xl m-0">{ greeting }</h1>
    </main>
  );
}
