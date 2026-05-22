import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { authAPI } from "../services/api.js";

export default function SSO() {
    const [params] = useSearchParams();
    const navigate = useNavigate();
    const [message, setMessage] = useState("Выполняется вход...");

    useEffect(() => {
        let cancelled = false;

        const run = async () => {
            const ticket = params.get("ticket");
            if (!ticket) {
                setMessage("SSO ticket не найден.");
                return;
            }

            try {
                const response = await authAPI.ssoExchange(ticket);
                const data = response.data;

                localStorage.setItem("token", data.token);
                localStorage.setItem(
                    "user",
                    JSON.stringify({
                        id: data.user_id,
                        email: data.email,
                        name: data.name,
                        surname: data.surname,
                        role: data.role,
                    })
                );

                if (data.application) {
                    localStorage.setItem("testingApplication", JSON.stringify(data.application));
                }

                if (cancelled) return;

                if ((data.role === "manager" || data.role === "admin") && data.next) {
                    navigate(data.next, { replace: true });
                    return;
                }

                const applicationId = data.application?.application?.id;
                if (data.role === "intern" && data.test_link) {
                    const query = applicationId ? `?application_id=${applicationId}` : "";
                    navigate(`/test/${data.test_link}${query}`, { replace: true });
                    return;
                }

                navigate(data.role === "intern" ? "/myTestStudent" : "/tests", { replace: true });
            } catch (error) {
                console.error("SSO exchange failed", error);
                setMessage("Не удалось выполнить вход через CRM.");
            }
        };

        run();
        return () => {
            cancelled = true;
        };
    }, [navigate, params]);

    return <div style={{ padding: 32 }}>{message}</div>;
}
