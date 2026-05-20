import { HomeEventsShowcase } from "@/features/home/home-events-showcase";
import { Footer } from "@/widgets/footer/footer";
import { Header } from "@/widgets/header/header";

export default function HomePage() {
  return (
    <div className="flex min-h-screen flex-col bg-bg">
      <Header variant="auto" />
      <HomeEventsShowcase />
      <Footer />
    </div>
  );
}
