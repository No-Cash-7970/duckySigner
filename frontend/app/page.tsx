'use client';

import Link from "next/link";
import { useState, useEffect } from "react";

export default function Home() {
  const [greeting, setGreeting] = useState('');
  const [time, setTime] = useState('');

  // Using Wails bindings must be done in a `useEffect` because they need the global `window` object
  useEffect(() => {
    (async () =>  {
      // Wails bindings imports must be dynamic also because importing bindings relies on the
      // `window` object too
      const Greet = (await import("@/bindings/main/GreetService")).Greet;
      // Output values from the Wails bindings to be displayed should be placed into in some kind of
      // state variable
      setGreeting(await Greet('user'));

      // Even Wails events need to be in a `useEffect`
      const Events = (await import("@wailsio/runtime")).Events;
      Events.On('time', (time: { name: string, data: string }) => {
          setTime(time.data);
      });
    })();
  }, []);

  return (
    <main className="prose bg-base h-[100vh] max-w-none grid place-content-center">
      <div className="text-center">{time}</div>
      <h1 className="text-5xl m-0">{ greeting }</h1>
      <div className="not-prose mt-6 flex justify-center">
        <ul className="menu menu-vertical sm:menu-horizontal bg-base-200 rounded-box">
          <li><Link href='/wallets'>Wallets List</Link></li>
          <li><a>Item 2</a></li>
          <li><a>Item 3</a></li>
        </ul>
      </div>
    </main>
  );
}
