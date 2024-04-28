<script lang="ts">
  import { Greet } from "$lib/wails-bindings/services/GreetService";
  import { Events } from "@wailsio/runtime";
  import { onDestroy } from "svelte";

  const greet = Greet('user');

  let currentTime: string;
  const unregTimeEvt = Events.On('time', (time: { name: string, data: string }) => {
    currentTime = time.data;
    console.log(currentTime);
  });

  onDestroy(() => {
    unregTimeEvt();
  });

</script>

<div class="text-center">{currentTime}</div>
<h1 class="text-5xl m-0">
  {#await greet then greeting }
    {greeting}
  {/await}
</h1>
<div class="not-prose mt-6 flex justify-center">
  <ul class="menu menu-vertical sm:menu-horizontal bg-base-200 rounded-box">
    <li><a href="/wallets">Wallets List</a></li>
    <li><a href="#item2">Item 2</a></li>
    <li><a href="#item3">Item 3</a></li>
  </ul>
</div>
