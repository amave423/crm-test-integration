import { useState, useEffect } from "react";
import {
    DndContext,
    closestCenter,
    KeyboardSensor,
    PointerSensor,
    useSensor,
    useSensors,
} from "@dnd-kit/core";
import {
    arrayMove,
    SortableContext,
    sortableKeyboardCoordinates,
    verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import EditIcon from "../assets/edit.svg?react";

import { useNavigate, useLocation } from "react-router-dom";
import SortableQuestion from "../components/Question";
import PassingCriteria from "../components/questions/PassingCriteria.jsx";
import ResultMessages from "../components/questions/ResultMessages";
import "../styles/createTest.css";
import LogoutButton from "../components/LogoutButton.jsx";
import TimeBox from "../components/details/TimeBox.jsx";
import BackIcon from "../assets/back.svg?react";
import { API_BASE_URL } from "../services/api.js";


function useAppSensors() {
    const pointerSensor = useSensor(PointerSensor);
    const keyboardSensor = useSensor(KeyboardSensor, {
        coordinateGetter: sortableKeyboardCoordinates,
    });

    return useSensors(pointerSensor, keyboardSensor);
}

export default function CreateTest() {
    const navigate = useNavigate();
    const location = useLocation();

    const isEditing = location.state?.editing || false;

    const storedEditingTest = (() => {
        try {
            const raw = localStorage.getItem("editingTest");
            return raw ? JSON.parse(raw) : null;
        } catch {
            return null;
        }
    })();

    const editingTest =
        (location.state?.test && location.state.test.questions
            ? location.state.test
            : storedEditingTest) || null;

    const deleteOnSave = location.state?.deleteOnSave || false;


    const [title, setTitle] = useState(
        isEditing ? editingTest?.title || "" : ""
    );

    const [description, setDescription] = useState(
        isEditing ? editingTest?.description || "" : ""
    );
    const [showQuestionMenu, setShowQuestionMenu] = useState(false);


    const [time, setTime] = useState(
        isEditing && editingTest?.complete_time
            ? {
                hours: Math.floor(editingTest.complete_time / 3600),
                minutes: Math.floor(
                    (editingTest.complete_time % 3600) / 60
                ),
                seconds: editingTest.complete_time % 60,
            }
            : {
                hours: 0,
                minutes: 0,
                seconds: 0,
            }
    );

    const [passingCriteria, setPassingCriteria] = useState(
        isEditing
            ? {
                type: editingTest?.is_percentage ? "percentage" : "points",
                percentage: editingTest?.is_percentage
                    ? editingTest.threshold
                    : 75,
                points: editingTest?.is_percentage
                    ? 0
                    : editingTest.threshold,
            }
            : {
                type: "percentage",
                percentage: 75,
                points: 0,
            }
    );

    const [resultMessages, setResultMessages] = useState(
        isEditing
            ? {
                success: editingTest?.success_text || "",
                failure: editingTest?.fail_text || "",
            }
            : {
                success: "",
                failure: "",
            }
    );

    const [questions, setQuestions] = useState(() => {
        if (isEditing && editingTest && Array.isArray(editingTest.questions)) {
            return editingTest.questions.map((q, idx) => {
                const base = {
                    id: `q-${idx}-${Date.now()}`,
                    order: idx + 1,
                    type: q.type || "shortText",
                    text: q.text || "",
                    maxScore: q.maxScore || q.points || 15,
                };

                switch (base.type) {
                    case "shortText":
                        return {
                            ...base,
                            correctAnswers:
                                q.correctAnswers ||
                                q.correct_input ||
                                [""],
                            caseSensitive:
                                q.caseSensitive !== undefined
                                    ? q.caseSensitive
                                    : q.case_sensitive || false,
                        };
                    case "singleChoice":
                        return {
                            ...base,
                            options:
                                q.options ||
                                q.choice ||
                                [{ text: "", isCorrect: false }],
                        };
                    case "multipleChoice":
                        return {
                            ...base,
                            options:
                                q.options ||
                                q.choice ||
                                [{ text: "", isCorrect: false }],
                            scoringType: q.scoringType || "allOrNothing",
                        };
                    case "matching":
                        return {
                            ...base,
                            rows:
                                q.rows ||
                                q.matching ||
                                [{ option: "", answer: "" }],
                        };
                    case "ordering":
                        return {
                            ...base,
                            items:
                                q.items ||
                                q.sequence ||
                                [{ text: "" }],
                        };
                    default:
                        return {
                            ...base,
                            correctAnswers: [""],
                            caseSensitive: false,
                        };
                }
            });
        }

        return [
            {
                id: "1",
                order: 1,
                type: "shortText",
                text: "",
                correctAnswers: [""],
                caseSensitive: false,
                maxScore: 15,
            },
        ];
    });

    const calculateTotalPoints = () => {
        return questions.reduce((sum, q) => sum + (q.maxScore || 0), 0);
    };

    const calculateCompleteTime = () => {
        return time.hours * 3600 + time.minutes * 60 + time.seconds;
    };

    const sensors = useAppSensors();

    const addQuestion = (type) => {
        const baseQuestion = {
            id: Date.now().toString(),
            order: questions.length + 1,
            type,
            text: "",
            maxScore: 15,
        };

        switch (type) {
            case "shortText":
                baseQuestion.correctAnswers = [""];
                baseQuestion.caseSensitive = false;
                break;
            case "singleChoice":
                baseQuestion.options = [{ text: "", isCorrect: false }];
                break;
            case "multipleChoice":
                baseQuestion.options = [{ text: "", isCorrect: false }];
                baseQuestion.scoringType = "allOrNothing";
                break;
            case "matching":
                baseQuestion.rows = [{ option: "", answer: "" }];
                break;
            case "ordering":
                baseQuestion.items = [{ text: "" }];
                break;
        }

        setQuestions([...questions, baseQuestion]);
    };

    const updateQuestion = (id, field, value) => {
        setQuestions(
            questions.map((q) =>
                q.id === id ? { ...q, [field]: value } : q
            )
        );
    };

    const deleteQuestion = (id) => {
        setQuestions(questions.filter((q) => q.id !== id));
    };

    const handleDragEnd = (event) => {
        const { active, over } = event;

        if (over && active.id !== over.id) {
            setQuestions((items) => {
                const oldIndex = items.findIndex(
                    (item) => item.id === active.id
                );
                const newIndex = items.findIndex(
                    (item) => item.id === over.id
                );
                const newItems = arrayMove(items, oldIndex, newIndex);

                return newItems.map((item, idx) => ({
                    ...item,
                    order: idx + 1,
                }));
            });
        }
    };

    const handleSave = async () => {
        if (!title.trim()) {
            alert("Р’РІРµРґРёС‚Рµ РЅР°Р·РІР°РЅРёРµ С‚РµСЃС‚Р°!");
            return;
        }

        const testData = {
            title: title.trim(),
            description: description.trim(),
            is_percentage: passingCriteria.type === "percentage",
            fail_text: resultMessages.failure || "",
            success_text: resultMessages.success || "",
            complete_time: calculateCompleteTime() || 3600,
            threshold:
                passingCriteria.type === "percentage"
                    ? passingCriteria.percentage
                    : passingCriteria.points,
            questions: questions.map((q, index) => {
                let options = {};
                let questionType;

                switch (q.type) {
                    case "singleChoice":
                        questionType = "single_choice";
                        options = {
                            choice:
                                q.options?.map((opt) => ({
                                    text: opt.text,
                                    is_true: opt.isCorrect,
                                })) || [],
                        };
                        break;
                    case "multipleChoice":
                        questionType = "multiple_choice";
                        options = {
                            choice:
                                q.options?.map((opt) => ({
                                    text: opt.text,
                                    is_true: opt.isCorrect,
                                })) || [],
                        };
                        break;
                    case "shortText":
                        questionType = "text_input";
                        options = {
                            correct_input: q.correctAnswers || [],
                            case_sensitive: q.caseSensitive || false,
                        };
                        break;
                    case "matching":
                        questionType = "matching";
                        options = {
                            matching:
                                q.rows?.map((row) => ({
                                    left: row.option,
                                    right: row.answer,
                                })) || [],
                        };
                        break;
                    case "ordering":
                        questionType = "correct_order";
                        options = {
                            sequence:
                                q.items?.map((item, itemIdx) => ({
                                    text: item.text,
                                    order: itemIdx + 1,
                                })) || [],
                        };
                        break;
                    default:
                        questionType = "text_input";
                }

                return {
                    text: q.text || "",
                    points: q.maxScore || 0,
                    type: questionType,
                    order_number: index + 1,
                    options: options,
                };
            }),
        };

        console.log(
            "РћС‚РїСЂР°РІР»СЏРµРјС‹Рµ РґР°РЅРЅС‹Рµ РЅР° Р±СЌРєРµРЅРґ:",
            JSON.stringify(testData, null, 2)
        );

        try {
            const token = localStorage.getItem("token");
            if (!token) {
                alert("РўСЂРµР±СѓРµС‚СЃСЏ Р°РІС‚РѕСЂРёР·Р°С†РёСЏ!");
                navigate("/login");
                return;
            }

            if (isEditing && deleteOnSave && editingTest?.id) {
                const testId =
                    editingTest.ID || editingTest.id || editingTest.Id;
                if (testId) {
                    console.log(
                        `РЈРґР°Р»СЏРµРј СЃС‚Р°СЂС‹Р№ С‚РµСЃС‚ СЃ ID: ${testId} РїРµСЂРµРґ СЃРѕР·РґР°РЅРёРµРј РЅРѕРІРѕРіРѕ`
                    );

                    const deleteResponse = await fetch(
                        `${API_BASE_URL}/api/manager/tests/delete/${testId}`,
                        {
                            method: "POST",
                            headers: {
                                Authorization: `Bearer ${token}`,
                                "Content-Type": "application/json",
                            },
                        }
                    );

                    const deleteResponseText =
                        await deleteResponse.text();
                    console.log(
                        "РћС‚РІРµС‚ РїСЂРё СѓРґР°Р»РµРЅРёРё СЃС‚Р°СЂРѕРіРѕ С‚РµСЃС‚Р°:",
                        deleteResponseText
                    );

                    if (!deleteResponse.ok) {
                        console.error(
                            "РќРµ СѓРґР°Р»РѕСЃСЊ СѓРґР°Р»РёС‚СЊ СЃС‚Р°СЂС‹Р№ С‚РµСЃС‚. РЎРѕР·РґР°РµРј РЅРѕРІС‹Р№ С‚РµСЃС‚ РїРѕРІРµСЂС… СЃСѓС‰РµСЃС‚РІСѓСЋС‰РµРіРѕ."
                        );
                    } else {
                        console.log("РЎС‚Р°СЂС‹Р№ С‚РµСЃС‚ СѓСЃРїРµС€РЅРѕ СѓРґР°Р»РµРЅ");
                    }
                }
            }

            const response = await fetch(
                `${API_BASE_URL}/api/manager/tests`,
                {
                    method: "POST",
                    headers: {
                        Authorization: `Bearer ${token}`,
                        "Content-Type": "application/json",
                    },
                    body: JSON.stringify(testData),
                }
            );

            const responseText = await response.text();

            if (!response.ok) {
                console.error("РћС‚РІРµС‚ СЃРµСЂРІРµСЂР° (С‚РµРєСЃС‚):", responseText);
                console.error("РЎС‚Р°С‚СѓСЃ РѕС€РёР±РєРё:", response.status);
                throw new Error(`РћС€РёР±РєР° HTTP: ${response.status}`);
            }

            let result;
            try {
                result = JSON.parse(responseText);
            } catch (e) {
                console.error(
                    "РќРµ СѓРґР°Р»РѕСЃСЊ СЂР°СЃРїР°СЂСЃРёС‚СЊ JSON РѕС‚РІРµС‚:",
                    responseText
                );
                throw new Error("РЎРµСЂРІРµСЂ РІРµСЂРЅСѓР» РЅРµРєРѕСЂСЂРµРєС‚РЅС‹Р№ JSON");
            }

            console.log("РЈСЃРїРµС€РЅС‹Р№ РѕС‚РІРµС‚ РѕС‚ СЃРµСЂРІРµСЂР°:", result);


            console.log("РЈСЃРїРµС€РЅС‹Р№ РѕС‚РІРµС‚ РѕС‚ СЃРµСЂРІРµСЂР°:", result);

            const savedId = result?.id || result?.test_id || editingTest?.id;
            if (savedId) {
                try {
                    const extendedTest = {
                        ...(editingTest || {}),
                        id: savedId,
                        title: testData.title,
                        description: testData.description,
                        is_percentage: testData.is_percentage,
                        threshold: testData.threshold,
                        success_text: testData.success_text,
                        fail_text: testData.fail_text,
                        complete_time: testData.complete_time,
                        questions,
                    };

                    const raw = localStorage.getItem("savedTestsExtended");
                    const list = raw ? JSON.parse(raw) : [];

                    const filtered = Array.isArray(list)
                        ? list.filter((t) => t.id !== savedId)
                        : [];

                    filtered.push(extendedTest);
                    localStorage.setItem(
                        "savedTestsExtended",
                        JSON.stringify(filtered)
                    );
                } catch (e) {
                    console.error("РќРµ СѓРґР°Р»РѕСЃСЊ СЃРѕС…СЂР°РЅРёС‚СЊ Р»РѕРєР°Р»СЊРЅС‹Р№ С‚РµСЃС‚ СЃ РІРѕРїСЂРѕСЃР°РјРё", e);
                }
            }


            localStorage.removeItem("editingTest");

            alert(
                isEditing
                    ? "РўРµСЃС‚ СѓСЃРїРµС€РЅРѕ РѕР±РЅРѕРІР»РµРЅ!"
                    : "РўРµСЃС‚ СѓСЃРїРµС€РЅРѕ СЃРѕР·РґР°РЅ РЅР° СЃРµСЂРІРµСЂРµ!"
            );
            navigate("/tests", { replace: true });

        } catch (error) {
            console.error("РћС€РёР±РєР° РїСЂРё СЃРѕР·РґР°РЅРёРё С‚РµСЃС‚Р°:", error);
            alert(
                `РќРµ СѓРґР°Р»РѕСЃСЊ ${
                    isEditing ? "РѕР±РЅРѕРІРёС‚СЊ" : "СЃРѕР·РґР°С‚СЊ"
                } С‚РµСЃС‚ РЅР° СЃРµСЂРІРµСЂРµ: ${error.message}\n\nРџСЂРѕРІРµСЂСЊС‚Рµ РєРѕРЅСЃРѕР»СЊ РґР»СЏ РґРµС‚Р°Р»РµР№.`
            );
        }
    };

    const handleBack = () => {
        navigate("/tests");
    };

    const questionTypes = [
        { key: "shortText", label: "Р—Р°РґР°РЅРёСЏ РЅР° СЂСѓС‡РЅРѕР№ РІРІРѕРґ" },
        { key: "singleChoice", label: "РћРґРёРЅРѕС‡РЅС‹Р№ РІС‹Р±РѕСЂ" },
        { key: "multipleChoice", label: "РњРЅРѕР¶РµСЃС‚РІРµРЅРЅС‹Р№ РІС‹Р±РѕСЂ" },
        { key: "matching", label: "РќР° СЃРѕРѕС‚РЅРѕС€РµРЅРёРµ" },
        {
            key: "ordering",
            label: "РќР° СЂР°СЃРїРѕР»РѕР¶РµРЅРёРµ РІ РїСЂР°РІРёР»СЊРЅРѕРј РїРѕСЂСЏРґРєРµ",
        },
    ];

    return (
        <div className="tests-page">
            <div
                className="test-page"
                style={{ position: "absolute", left: "1430px", top: "0px" }}
            >
                <LogoutButton />
            </div>

            <div className="create-wrapper">
                <div className="create-left">
                    <div className="stat-top-bar2">
                        <button className="stat-back-btn2" onClick={handleBack}>
                            <BackIcon />
                        </button>
                        <h1>РЎРѕР·РґР°РЅРёРµ С‚РµСЃС‚Р°</h1>

                    </div>
                    <div className="tests-line"></div>
                    <div className="title-input-container">
                        <input
                            className="test-desk1"
                            placeholder="РќР°Р·РІР°РЅРёРµ"
                            value={title}
                            onChange={(e) => setTitle(e.target.value)}
                        />
                        <EditIcon />
                    </div>
                    <div className="title-input-container-desk">
                    <input
                        className="test-desk"
                        placeholder="РћРїРёСЃР°РЅРёРµ С‚РµСЃС‚Р°"
                        value={description}
                        onChange={(e) =>
                            setDescription(e.target.value)
                        }
                    />
                        <EditIcon />
                    </div>
                    {/*
                    <PassingCriteria
                        criteria={passingCriteria}
                        updateCriteria={setPassingCriteria}
                        totalPoints={calculateTotalPoints()}
                    />
                       */}
                    <DndContext
                        sensors={sensors}
                        collisionDetection={closestCenter}
                        onDragEnd={handleDragEnd}
                    >
                        <SortableContext
                            items={questions.map((q) => q.id)}
                            strategy={verticalListSortingStrategy}
                        >
                            {questions.map((question) => (
                                <SortableQuestion
                                    key={question.id}
                                    question={question}
                                    updateQuestion={updateQuestion}
                                    deleteQuestion={deleteQuestion}
                                />
                            ))}
                        </SortableContext>
                    </DndContext>

                    {/* <ResultMessages
                        messages={resultMessages}
                        updateMessages={setResultMessages}
                    />
                    */}
                </div>

                <div className="create-right">
                    <div className="create-right-inner">
                        <button
                            className="save-btn"
                            onClick={handleSave}
                        >
                            {isEditing
                                ? "РЎРѕС…СЂР°РЅРёС‚СЊ РёР·РјРµРЅРµРЅРёСЏ"
                                : "РЎРѕР·РґР°С‚СЊ С‚РµСЃС‚"}
                        </button>
                        <h3>РџРѕР»СЏ С‚РµСЃС‚Р°</h3>
                        <div className="right-section">
                            <button
                                className="right-btn-toggle"
                                onClick={() => setShowQuestionMenu(!showQuestionMenu)}
                            >
                                Р”РѕР±Р°РІРёС‚СЊ РЅРѕРІС‹Р№ РІРѕРїСЂРѕСЃ
                                <span className={`toggle-arrow ${showQuestionMenu ? 'open' : ''}`}>в–ј</span>
                            </button>

                            {showQuestionMenu && questionTypes.map((type) => (
                                <button
                                    key={type.key}
                                    className="right-btn"
                                    onClick={() => {
                                        addQuestion(type.key);
                                    }}
                                >
                                    {type.label}
                                </button>
                            ))}
                        </div>
                        {/* <TimeBox time={time} setTime={setTime} />
        */}
                    </div>
                </div>

            </div>
        </div>
    );
}


